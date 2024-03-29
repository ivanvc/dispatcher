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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/ivanvc/dispatcher/pkg/api/v1beta1"
)

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

// Implements conversion to v1beta1.
func (j *JobTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.JobTemplate)
	dst.ObjectMeta = j.ObjectMeta
	dst.Spec.JobTemplateSpec = j.Spec.JobTemplateSpec

	return nil
}

// Converts from the Hub version (v1beta1) to this version.
func (j *JobTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.JobTemplate)
	j.ObjectMeta = src.ObjectMeta
	j.Spec.JobTemplateSpec = src.Spec.JobTemplateSpec

	return nil
}

func init() {
	SchemeBuilder.Register(&JobTemplate{}, &JobTemplateList{})
}
