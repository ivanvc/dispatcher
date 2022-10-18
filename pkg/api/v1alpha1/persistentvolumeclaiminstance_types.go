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

// PersistentVolumeClaimInstanceSpec defines the desired state of PersistentVolumeClaimInstance
type PersistentVolumeClaimInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	// The JobTemplate from where to read the PersistentVolumeClaimSpec from.
	JobTemplateName string `json:"jobTemplateName"`

	//+kubebuilder:validation:Required
	// A UUID for the execution.
	UUID string `json:"uuid"`

	//+kubebuilder:validation:Required
	// The Timestamp of when the execution was created.
	Timestamp metav1.Time `json:"timestamp"`
}

// PersistentVolumeClaimInstanceStatus defines the observed state of PersistentVolumeClaimInstance
type PersistentVolumeClaimInstanceStatus struct {
	// Phase has the current state of the PersistentVolumeClaim, it could be one of:
	// - "Invalid": The PersistentVolumeClaimInstance is referring to a non-existent JobTemplate;
	// - "Waiting": Waiting to be created;
	// - "Created": The PVC was created;
	Phase PersistentVolumeClaimInstancePhase `json:"phase"`

	// PersistentVolumeClaim holds the actual claim.
	// +optional
	PersistentVolumeClaim corev1.ObjectReference `json:"job,omitempty"`
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

// PersistentVolumeClaimInstancePhase describes the current state of the PVC.
type PersistentVolumeClaimInstancePhase string

const (
	// JobTemplate does not exist.
	PersistentVolumeClaimInstanceInvalidPhase PersistentVolumeClaimInstancePhase = "Invalid"
	// Waiting for the PVC be created.
	PersistentVolumeClaimInstanceWaitingPhase PersistentVolumeClaimInstancePhase = "Waiting"
	// The PVC was created.
	PersistentVolumeClaimInstanceCreatedPhase PersistentVolumeClaimInstancePhase = "Created"
)
