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

package v1beta1

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
	// The JobTemplate to execute.
	JobTemplateName string `json:"jobTemplateName"`

	//+optional
	// The execution arguments to pass to the JobTemplate's Job.
	Payload string `json:"payload,omitempty"`
}

// JobExecutionStatus defines the observed state of JobExecution
type JobExecutionStatus struct {
	// Represents the observations of a JobExecution's state. The state of the JobExecution is tied to the Job it manages.
	// Conditions.type are: "Waiting", "Running", "Succeeded".
	// Conditions.status are one of True, False, Unknown.
	// Conditions.reason defines a camelCase expected values and meanings for this field.
	// Conditions.Message is a human readable message indicating details about the transition.

	// Conditions store the status conditions of a JobExecution.
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Job has a reference to the Job from this execution.
	// +optional
	Job corev1.ObjectReference `json:"job,omitempty"`
}

// JobExecutionConditionType describes the observed state of a JobExecution and its Job.
type JobExecutionConditionType string

const (
	JobExecutionWaiting   JobExecutionConditionType = "Waiting"
	JobExecutionRunning   JobExecutionConditionType = "Running"
	JobExecutionSucceeded JobExecutionConditionType = "Succeeded"
)

//+kubebuilder:storageversion
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

// Set v1beta1 as the Hub.
func (*JobExecution) Hub() {}

func init() {
	SchemeBuilder.Register(&JobExecution{}, &JobExecutionList{})
}
