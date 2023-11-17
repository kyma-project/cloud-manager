/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NfsVolumeSpec defines the desired state of NfsVolume
type NfsVolumeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Capacity      Capacity      `json:"capacity"`
	ReclaimPolicy ReclaimPolicy `json:"reclaimPolicy"`
}

type Capacity struct {
	Storage string `json:"storage"`
}

type ReclaimPolicy string

const (
	ReclaimPolicyRetain  ReclaimPolicy = "Retain"
	ReclaimPolicyRecycle ReclaimPolicy = "Recycle"
)

// NfsVolumeStatus defines the observed state of NfsVolume
type NfsVolumeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase                    Phase       `json:"phase"`
	Error                    string      `json:"error,omitempty"`
	NfsCreationTimestamp     metav1.Time `json:"nfsCreationTimestamp,omitempty"`
	NfsLastModifiedTimestamp metav1.Time `json:"nfsLastModifiedTimestamp,omitempty"`
}

type Phase string

const (
	PhaseCreating Phase = "Creating"
	PhaseUpdating Phase = "Updating"
	PhaseDeleting Phase = "Deleting"
	PhaseReady    Phase = "Ready"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NfsVolume is the Schema for the nfsvolumes API
type NfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsVolumeSpec   `json:"spec,omitempty"`
	Status NfsVolumeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NfsVolumeList contains a list of NfsVolume
type NfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NfsVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NfsVolume{}, &NfsVolumeList{})
}
