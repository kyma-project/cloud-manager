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
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CceeNfsVolumeSpec defines the desired state of CceeNfsVolume
type CceeNfsVolumeSpec struct {
	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	CapacityGb int `json:"capacityGb"`

	// +optional
	PersistentVolume *NameLabelsAnnotationsSpec `json:"volume,omitempty"`

	// +optional
	PersistentVolumeClaim *NameLabelsAnnotationsSpec `json:"volumeClaim,omitempty"`
}

type NameLabelsAnnotationsSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// CceeNfsVolumeStatus defines the observed state of CceeNfsVolume
type CceeNfsVolumeStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	Server string `json:"server,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CceeNfsVolume is the Schema for the cceenfsvolumes API
type CceeNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CceeNfsVolumeSpec   `json:"spec,omitempty"`
	Status CceeNfsVolumeStatus `json:"status,omitempty"`
}

func (in *CceeNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *CceeNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *CceeNfsVolume) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfs
}

func (in *CceeNfsVolume) SpecificToProviders() []string {
	return []string{"openstack"}
}

func (in *CceeNfsVolume) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *CceeNfsVolume) State() string {
	return in.Status.State
}

func (in *CceeNfsVolume) SetState(v string) {
	in.Status.State = v
}

func (in *CceeNfsVolume) CloneForPatchStatus() client.Object {
	return &CceeNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CceeNfsVolume",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

// +kubebuilder:object:root=true

// CceeNfsVolumeList contains a list of CceeNfsVolume
type CceeNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CceeNfsVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CceeNfsVolume{}, &CceeNfsVolumeList{})
}
