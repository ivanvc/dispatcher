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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// JobTemplate is the Schema for the jobtemplate API
type JobTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JobTemplateSpec `json:"spec"`
}

// JobTemplateSpec defines the desired state of JobTemplate
type JobTemplateSpec struct {
	// Specifies the Job that will be created when executing the Job.
	batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// +kubebuilder:object:root=true
// JobTemplateList contains a list of JobTemplate
type JobTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobTemplate `json:"items"`
}

// Set v1beta1 as the Hub.
func (*JobTemplate) Hub() {}

func init() {
	SchemeBuilder.Register(&JobTemplate{}, &JobTemplateList{})
}
