/*
Copyright 2022 Ivan Valdes.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	dispatcherv1alpha1 "github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	"github.com/ivanvc/dispatcher/pkg/template"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	jobExecutionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_executions_total",
		Help: "The total number of dispatched JobExecutions.",
	})
	jobExecutionsFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_executions_failures_total",
		Help: "The total number of failed JobExecutions.",
	})
	jobExecutionsSuccessTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_executions_success_total",
		Help: "The total number of successful JobExecutions.",
	})
)

func init() {
	metrics.Registry.MustRegister(
		jobExecutionsTotal,
		jobExecutionsFailuresTotal,
		jobExecutionsSuccessTotal,
	)
}

// JobExecutionReconciler reconciles a JobExecution object
type JobExecutionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dispatcher.ivan.vc,resources=jobexecutions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dispatcher.ivan.vc,resources=jobtemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=dispatcher.ivan.vc,resources=jobexecutions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dispatcher.ivan.vc,resources=jobexecutions/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *JobExecutionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	je := new(dispatcherv1alpha1.JobExecution)
	if err := r.Get(ctx, req.NamespacedName, je); err != nil {
		if errors.IsNotFound(err) {
			log.Info("JobExecution resource not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get JobExecution, requeueing")
		return ctrl.Result{}, err
	}

	jt, err := r.getJobTemplate(ctx, je)
	if err != nil {
		log.Error(err, "Failed to get JobTemplate, requeueing.")
		return ctrl.Result{}, err
	}

	job, err := r.getJob(ctx, je)
	if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	if job == nil {
		if je.Status.Phase == v1alpha1.JobExecutionCompletedPhase {
			log.Info("JobExecution is already completed", "JobExecution", je.ObjectMeta.Name)
			if err := r.Delete(ctx, je); err != nil {
				return ctrl.Result{}, err
			}

			jobExecutionsSuccessTotal.Inc()
			return ctrl.Result{}, nil
		}

		if je.Status.Phase == v1alpha1.JobExecutionFailedPhase {
			log.Info("JobExecution failed to complete", "JobExecution", je.ObjectMeta.Name)
			if err := r.Delete(ctx, je); err != nil {
				return ctrl.Result{}, err
			}

			jobExecutionsFailuresTotal.Inc()
			return ctrl.Result{}, nil
		}

		if err := r.createJob(ctx, je, jt); err != nil {
			log.Error(err, "Error generating Job")
			return ctrl.Result{}, err
		}

		log.Info("Created Job, requeueing")
		jobExecutionsTotal.Inc()
		return ctrl.Result{Requeue: true}, nil
	}

	if job.Status.CompletionTime != nil {
		je.Status.Phase = v1alpha1.JobExecutionCompletedPhase
	} else if len(job.Status.Conditions) > 0 && hasFailedCondition(job) {
		je.Status.Phase = v1alpha1.JobExecutionFailedPhase
	} else if job.Status.StartTime != nil {
		je.Status.Phase = v1alpha1.JobExecutionActivePhase
	} else {
		je.Status.Phase = v1alpha1.JobExecutionWaitingPhase
	}

	jobRef, err := ref.GetReference(r.Scheme, job)
	if err != nil {
		log.Error(err, "Unable to make reference to job", "job", job)
		return ctrl.Result{}, err
	}
	je.Status.Job = *jobRef

	if err := r.Status().Update(ctx, je); err != nil {
		log.Error(err, "Failed to update JobExecution status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 15}, nil
}

// Returns true if there is at least one condition from the job that has the failed status.
func hasFailedCondition(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dispatcherv1alpha1.JobExecution{}).
		Complete(r)
}

// Generates a Job from a JobTemplate, by applying JobExecution's fields.
func (r *JobExecutionReconciler) generateJobFromTemplate(jobExecution *dispatcherv1alpha1.JobExecution, jobTemplate *dispatcherv1alpha1.JobTemplate) (*batchv1.Job, error) {
	jobTpl, err := template.BuildJob(&jobTemplate.Spec.JobTemplateSpec, jobExecution)
	if err != nil {
		return nil, err
	}
	job := &batchv1.Job{
		ObjectMeta: jobTpl.ObjectMeta,
		Spec:       jobTpl.Spec,
	}

	job.ObjectMeta.Namespace = jobTemplate.Namespace
	if len(job.ObjectMeta.Name) == 0 && len(job.ObjectMeta.GenerateName) == 0 {
		job.ObjectMeta.GenerateName = jobExecution.ObjectMeta.Name + "-"
	}
	if job.ObjectMeta.Labels == nil {
		job.ObjectMeta.Labels = make(map[string]string)
	}
	job.ObjectMeta.Labels["controller-uid"] = string(jobExecution.ObjectMeta.UID)
	job.ObjectMeta.Labels["job-execution-name"] = jobExecution.ObjectMeta.Name

	ctrl.SetControllerReference(jobExecution, job, r.Scheme)
	return job, nil
}

func (r *JobExecutionReconciler) getJobTemplate(ctx context.Context, jobExecution *dispatcherv1alpha1.JobExecution) (*dispatcherv1alpha1.JobTemplate, error) {
	jt := new(dispatcherv1alpha1.JobTemplate)
	if err := r.Get(ctx, types.NamespacedName{
		Name:      jobExecution.Spec.JobTemplateName,
		Namespace: jobExecution.Namespace,
	}, jt); err != nil {
		jobExecution.Status.Phase = v1alpha1.JobExecutionInvalidPhase
		if err := r.Status().Update(ctx, jobExecution); err != nil {
			return nil, err
		}
		return nil, err
	}
	return jt, nil
}

func (r *JobExecutionReconciler) getJob(ctx context.Context, jobExecution *dispatcherv1alpha1.JobExecution) (*batchv1.Job, error) {
	opts := []client.ListOption{
		client.InNamespace(jobExecution.Namespace),
		client.MatchingLabels{"controller-uid": string(jobExecution.ObjectMeta.UID)},
	}

	jobList := new(batchv1.JobList)
	if err := r.List(ctx, jobList, opts...); err != nil {
		return nil, err
	}

	if len(jobList.Items) == 0 {
		return nil, nil
	}

	return &jobList.Items[0], nil
}

func (r *JobExecutionReconciler) createJob(ctx context.Context, jobExecution *dispatcherv1alpha1.JobExecution, jobTemplate *dispatcherv1alpha1.JobTemplate) error {
	job, err := r.generateJobFromTemplate(jobExecution, jobTemplate)
	if err != nil {
		return err
	}
	if err := r.Create(ctx, job); err != nil {
		return err
	}
	return nil
}
