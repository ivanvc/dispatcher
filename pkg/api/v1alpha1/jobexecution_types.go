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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JobExecutionSpec defines the desired state of JobExecution
type JobExecutionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	// The Job Definition to execute.
	JobTemplateName string `json:"jobTemplateName"`
}

// JobExecutionStatus defines the observed state of JobExecution
type JobExecutionStatus struct {
	// State has the current state of the Job, it could be one of:
	// - "Invalid": The JobExecution is referring to a non-existent JobTemplate;
	// - "Waiting": Waiting to be scheduled;
	// - "Active": The Job is running;
	// - "Completed": The Job finished running.
	State JobExecutionState `json:"state"`

	// Job holds the actual Job that is executing.
	// +optional
	Job corev1.ObjectReference `json:"job,omitempty"`
}

// JobExecutionState describes the current state of the Job.
type JobExecutionState string

const (
	// JobTemplate does not exist.
	JobExecutionInvalidState JobExecutionState = "Invalid"
	// Job is waiting to be scheduled.
	JobExecutionWaitingState JobExecutionState = "Waiting"
	// Job is running.
	JobExecutionActiveState JobExecutionState = "Active"
	// Job has finished running.
	JobExecutionCompletedState JobExecutionState = "Completed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JobExecution is the Schema for the jobexecutions API
type JobExecution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobExecutionSpec   `json:"spec,omitempty"`
	Status JobExecutionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JobExecutionList contains a list of JobExecution
type JobExecutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobExecution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobExecution{}, &JobExecutionList{})
}
