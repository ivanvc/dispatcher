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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	dispatcherv1alpha1 "github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	"github.com/ivanvc/dispatcher/pkg/template"
)

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
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JobExecution object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
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

	jt := new(dispatcherv1alpha1.JobTemplate)
	if err := r.Get(ctx, types.NamespacedName{
		Name:      je.Spec.JobTemplateName,
		Namespace: je.Namespace,
	}, jt); err != nil {
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get JobTemplate, requeueing.")
		je.Status.Phase = v1alpha1.JobExecutionInvalidPhase
		if err := r.Status().Update(ctx, je); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	opts := []client.ListOption{
		client.InNamespace(je.Namespace),
		client.MatchingLabels{"controller-uid": string(je.ObjectMeta.UID)},
	}

	if jt.Spec.PersistentVolumeClaimTemplateSpec.Spec.Resources.Size() > 0 {
		pvcList := new(corev1.PersistentVolumeClaimList)
		if err := r.List(ctx, pvcList, opts...); err != nil {
			log.Error(err, "Failed to get PVC")
			return ctrl.Result{}, err
		}

		if len(pvcList.Items) == 0 {
			pvc, err := r.generatePVCFromTemplate(jt, je)
			if err != nil {
				log.Error(err, "Error generating PVC")
			}
			log.Info("Creating PVC", "PVC.Namespace", pvc.Namespace)
			if err := r.Create(ctx, pvc); err != nil {
				log.Error(err, "Failed to create new PVC", "PVC.Namespace", pvc.Namespace)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	jobList := new(batchv1.JobList)
	if err := r.List(ctx, jobList, opts...); err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}
	if len(jobList.Items) == 0 {
		job, err := r.generateJobFromTemplate(jt, je)
		if err != nil {
			log.Error(err, "Error generating Job")
			return ctrl.Result{}, err
		}
		log.Info("Creating Job", "Job.Namespace", job.Namespace)
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	job := jobList.Items[0]
	if job.Status.CompletionTime != nil {
		je.Status.Phase = v1alpha1.JobExecutionCompletedPhase
	} else if job.Status.StartTime != nil {
		je.Status.Phase = v1alpha1.JobExecutionActivePhase
	} else {
		je.Status.Phase = v1alpha1.JobExecutionWaitingPhase
	}

	jobRef, err := ref.GetReference(r.Scheme, &job)
	if err != nil {
		log.Error(err, "Unable to make reference to job", "job", job)
		return ctrl.Result{}, err
	}
	je.Status.Job = *jobRef

	if err := r.Status().Update(ctx, je); err != nil {
		log.Error(err, "Failed to update JobExecution status")
		return ctrl.Result{}, err
	}

	if je.Status.Phase == v1alpha1.JobExecutionCompletedPhase {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: time.Second * 15}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dispatcherv1alpha1.JobExecution{}).
		Complete(r)
}

// Generates a Job from a JobTemplate, by applying JobExecution's fields.
func (r *JobExecutionReconciler) generateJobFromTemplate(jobTemplate *dispatcherv1alpha1.JobTemplate, jobExecution *dispatcherv1alpha1.JobExecution) (*batchv1.Job, error) {
	jobTpl, err := template.BuildJob(&jobTemplate.Spec.JobTemplateSpec, jobExecution)
	if err != nil {
		return nil, err
	}
	job := &batchv1.Job{
		ObjectMeta: jobTpl.ObjectMeta,
		Spec:       jobTpl.Spec,
	}

	job.ObjectMeta.Namespace = jobTemplate.Namespace
	if job.ObjectMeta.Labels == nil {
		job.ObjectMeta.Labels = make(map[string]string)
	}
	job.ObjectMeta.Labels["controller-uid"] = string(jobExecution.ObjectMeta.UID)
	job.ObjectMeta.Labels["job-execution-name"] = jobExecution.ObjectMeta.Name

	ctrl.SetControllerReference(jobExecution, job, r.Scheme)
	return job, nil
}

// Generates a PVC from a JobTemplate, by applying JobExecution's fields.
func (r *JobExecutionReconciler) generatePVCFromTemplate(jobTemplate *dispatcherv1alpha1.JobTemplate, jobExecution *dispatcherv1alpha1.JobExecution) (*corev1.PersistentVolumeClaim, error) {
	pvcTpl, err := template.BuildPersistentVolumeClaim(&jobTemplate.Spec.PersistentVolumeClaimTemplateSpec, jobExecution)
	if err != nil {
		return nil, err
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: pvcTpl.ObjectMeta,
		Spec:       pvcTpl.Spec,
	}

	pvc.ObjectMeta.Namespace = pvc.Namespace
	if pvc.ObjectMeta.Labels == nil {
		pvc.ObjectMeta.Labels = make(map[string]string)
	}
	pvc.ObjectMeta.Labels["controller-uid"] = string(jobExecution.ObjectMeta.UID)
	pvc.ObjectMeta.Labels["job-execution-name"] = jobExecution.ObjectMeta.Name

	ctrl.SetControllerReference(jobExecution, pvc, r.Scheme)
	return pvc, nil
}
