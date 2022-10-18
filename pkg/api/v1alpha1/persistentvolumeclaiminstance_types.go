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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PersistentVolumeClaimInstanceSpec defines the desired state of PersistentVolumeClaimInstance
type PersistentVolumeClaimInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PersistentVolumeClaimInstance. Edit persistentvolumeclaiminstance_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PersistentVolumeClaimInstanceStatus defines the observed state of PersistentVolumeClaimInstance
type PersistentVolumeClaimInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PersistentVolumeClaimInstance is the Schema for the persistentvolumeclaiminstances API
type PersistentVolumeClaimInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PersistentVolumeClaimInstanceSpec   `json:"spec,omitempty"`
	Status PersistentVolumeClaimInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PersistentVolumeClaimInstanceList contains a list of PersistentVolumeClaimInstance
type PersistentVolumeClaimInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PersistentVolumeClaimInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PersistentVolumeClaimInstance{}, &PersistentVolumeClaimInstanceList{})
}
