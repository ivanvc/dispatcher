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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	dispatcherv1beta1 "github.com/ivanvc/dispatcher/pkg/api/v1beta1"
	"github.com/ivanvc/dispatcher/pkg/template"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	waitingCondition   = string(dispatcherv1beta1.JobExecutionWaiting)
	runningCondition   = string(dispatcherv1beta1.JobExecutionRunning)
	succeededCondition = string(dispatcherv1beta1.JobExecutionSucceeded)
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
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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

	// Fetch the JobExecution
	je := new(dispatcherv1beta1.JobExecution)
	if err := r.Get(ctx, req.NamespacedName, je); err != nil {
		if errors.IsNotFound(err) {
			log.Info("JobExecution resource not found, ignoring as resouce must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get JobExecution, requeueing")
		return ctrl.Result{}, err
	}

	// If conditions are not set, initialize the slice
	if je.Status.Conditions == nil || len(je.Status.Conditions) == 0 {
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    waitingCondition,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		if err := r.Status().Update(ctx, je); err != nil {
			log.Error(err, "Failed to update JobExecution status")
			return ctrl.Result{}, err
		}

		// Re-fetch JobExecution to avoid having to re-run reocnciliation loop
		if err := r.Get(ctx, req.NamespacedName, je); err != nil {
			log.Error(err, "Failed to re-fetch JobExecution")
			return ctrl.Result{}, err
		}
	}

	jt, err := r.getJobTemplate(ctx, je)
	if err != nil {
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    waitingCondition,
			Status:  metav1.ConditionUnknown,
			Reason:  "FetchJobTemplateError",
			Message: "Failed fetching JobTemplate",
		})
		if err := r.Status().Update(ctx, je); err != nil {
			log.Error(err, "Failed to update JobExecution status")
			return ctrl.Result{}, err
		}

		log.Error(err, "Failed to get JobTemplate, requeueing")
		r.Recorder.Eventf(
			je,
			corev1.EventTypeWarning,
			"JobTemplateNotFound",
			"Failed fetching JobTemplate %s: %s",
			je.Spec.JobTemplateName,
			err.Error(),
		)
		return ctrl.Result{}, err
	}

	// Fetch owned Job
	job, err := r.getJob(ctx, je)
	if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	// If job is not found
	if job == nil {
		// If job is not running anymore, don't care about succeeded condition, as
		// it may or not finished successfully.
		if meta.IsStatusConditionFalse(je.Status.Conditions, runningCondition) {
			log.Info("JobExecution is already completed", "JobExecution", je.Name)
			if err := r.Delete(ctx, je); err != nil {
				log.Error(err, "Failed to delete JobExecution")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		// Create a job
		createdJob, err := r.createJob(ctx, je, jt)
		if err != nil {
			log.Error(err, "Error generating Job")
			return ctrl.Result{}, err
		}

		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    waitingCondition,
			Status:  metav1.ConditionTrue,
			Reason:  "JobCreated",
			Message: "Job created, waiting to be executed",
		})
		jobRef, err := ref.GetReference(r.Scheme, createdJob)
		if err != nil {
			log.Error(err, "Unable to make reference to job", "job", job)
			return ctrl.Result{}, err
		}
		je.Status.Job = *jobRef

		if err := r.Status().Update(ctx, je); err != nil {
			log.Error(err, "Failed to update JobExecution status")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(je, corev1.EventTypeNormal, "Created", "Job %s created", createdJob.Name)
		log.Info("Created Job, requeueing")
		jobExecutionsTotal.Inc()
		return ctrl.Result{Requeue: true}, nil
	}

	// Check status of JobExecution's owned Job
	if isJobStatusConditionTrue(job, batchv1.JobComplete) {
		r.Recorder.Eventf(je, corev1.EventTypeNormal, "Completed", "Job %s completed running", job.Name)
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    succeededCondition,
			Status:  metav1.ConditionTrue,
			Reason:  "JobSucceeded",
			Message: "Job ran successfully",
		})
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    runningCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "JobCompleted",
			Message: "Job completed running",
		})
		jobExecutionsSuccessTotal.Inc()
	} else if isJobStatusConditionTrue(job, batchv1.JobFailed) {
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    succeededCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "JobFailed",
			Message: "Job completed with a failed exit status",
		})
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    runningCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "JobCompleted",
			Message: "Job completed running",
		})
		r.Recorder.Eventf(je, corev1.EventTypeWarning, "Failed", "Job %s failed running", job.Name)
		jobExecutionsFailuresTotal.Inc()
	} else if job.Status.StartTime != nil {
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    waitingCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "JobRunning",
			Message: "Job is running",
		})
		meta.SetStatusCondition(&je.Status.Conditions, metav1.Condition{
			Type:    runningCondition,
			Status:  metav1.ConditionTrue,
			Reason:  "JobRunning",
			Message: "Job is running",
		})
		r.Recorder.Eventf(je, corev1.EventTypeNormal, "Started", "Job %s started running", job.Name)
	}

	if err := r.Status().Update(ctx, je); err != nil {
		log.Error(err, "Failed to update JobExecution status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 15}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dispatcherv1beta1.JobExecution{}).
		Complete(r)
}

// Generates a Job from a JobTemplate, by applying JobExecution's fields.
func (r *JobExecutionReconciler) generateJobFromTemplate(jobExecution *dispatcherv1beta1.JobExecution, jobTemplate *dispatcherv1beta1.JobTemplate) (*batchv1.Job, error) {
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
	job.ObjectMeta.Labels["controller-uid"] = string(jobExecution.GetUID())
	job.ObjectMeta.Labels["job-execution-name"] = jobExecution.Name

	ctrl.SetControllerReference(jobExecution, job, r.Scheme)
	return job, nil
}

// Gets the JobTemplate from a jobExecution.
func (r *JobExecutionReconciler) getJobTemplate(ctx context.Context, jobExecution *dispatcherv1beta1.JobExecution) (*dispatcherv1beta1.JobTemplate, error) {
	jt := new(dispatcherv1beta1.JobTemplate)
	if err := r.Get(ctx, types.NamespacedName{
		Name:      jobExecution.Spec.JobTemplateName,
		Namespace: jobExecution.Namespace,
	}, jt); err != nil {
		return nil, err
	}
	return jt, nil
}

// Gets the Job from a jobExecution
func (r *JobExecutionReconciler) getJob(ctx context.Context, jobExecution *dispatcherv1beta1.JobExecution) (*batchv1.Job, error) {
	if len(jobExecution.Status.Job.Name) == 0 {
		return nil, nil
	}
	job := new(batchv1.Job)
	if err := r.Get(ctx, types.NamespacedName{
		Name:      jobExecution.Status.Job.Name,
		Namespace: jobExecution.Namespace,
	}, job); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return job, nil
}

// Creates a Job from a jobExecution and its jobTemplate.
func (r *JobExecutionReconciler) createJob(ctx context.Context, jobExecution *dispatcherv1beta1.JobExecution, jobTemplate *dispatcherv1beta1.JobTemplate) (*batchv1.Job, error) {
	job, err := r.generateJobFromTemplate(jobExecution, jobTemplate)
	if err != nil {
		return nil, err
	}
	if err := r.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

// Returns true if the Job has a condition that matches the given status.
func isJobStatusConditionTrue(job *batchv1.Job, conditionType batchv1.JobConditionType) bool {
	if job.Status.Conditions == nil {
		return false
	}
	for _, condition := range job.Status.Conditions {
		if condition.Type == conditionType && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
