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

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	dispatcherv1alpha1 "github.com/ivanvc/dispatcher/api/v1alpha1"
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
	err := r.Get(ctx, req.NamespacedName, je)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("JobExecution resource not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get JobExecution, requeueing")
		return ctrl.Result{}, err
	}

	jt := new(dispatcherv1alpha1.JobTemplate)
	err = r.Get(ctx, types.NamespacedName{
		Name:      je.Spec.JobTemplateName,
		Namespace: je.Namespace,
	}, jt)
	if err != nil {
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get JobTemplate, requeueing.")
		je.Status.Phase = "Invalid JobTemplateName"
		if err := r.Status().Update(ctx, je); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	jobList := new(batchv1.JobList)
	opts := []client.ListOption{
		client.InNamespace(je.Namespace),
		client.MatchingLabels{"controller-uid": string(je.ObjectMeta.UID)},
	}

	err = r.List(ctx, jobList, opts...)
	if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}
	if len(jobList.Items) == 0 {
		job := r.jobFromDefinition(je, jt)
		log.Info("Creating Job", "Job.Namespace", job.Namespace)
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	je.Status.Phase = "Job created"
	je.Status.Job = jobList.Items[0].Name
	err = r.Status().Update(ctx, je)
	if err != nil {
		log.Error(err, "Failed to update JobExecution status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dispatcherv1alpha1.JobExecution{}).
		Complete(r)
}

// jobFromDefinition returns a Job from the JobTemplate.
func (r *JobExecutionReconciler) jobFromDefinition(je *dispatcherv1alpha1.JobExecution, jt *dispatcherv1alpha1.JobTemplate) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: jt.Spec.Template.ObjectMeta,
		Spec:       jt.Spec.Template.Spec,
	}
	job.ObjectMeta.GenerateName = jt.Name + "-"
	job.ObjectMeta.Namespace = jt.Namespace
	if job.ObjectMeta.Labels == nil {
		job.ObjectMeta.Labels = make(map[string]string)
	}
	job.ObjectMeta.Labels["controller-uid"] = string(je.ObjectMeta.UID)

	ctrl.SetControllerReference(je, job, r.Scheme)
	return job
}
