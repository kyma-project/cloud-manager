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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SapNfsVolumeSpec defines the desired state of SapNfsVolume
type SapNfsVolumeSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self > 0), message="The field capacityGb must be greater than zero"
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

// SapNfsVolumeStatus defines the observed state of SapNfsVolume
type SapNfsVolumeStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	Server string `json:"server,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`

	// +optional
	State string `json:"state,omitempty"`

	// Provisioned Capacity
	// +optional
	Capacity resource.Quantity `json:"capacity"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}

// SapNfsVolume is the Schema for the sapnfsvolumes API
type SapNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SapNfsVolumeSpec   `json:"spec,omitempty"`
	Status SapNfsVolumeStatus `json:"status,omitempty"`
}

func (in *SapNfsVolume) GetPVName() string {
	if in.Spec.PersistentVolume != nil && in.Spec.PersistentVolume.Name != "" {
		return in.Spec.PersistentVolume.Name
	}
	return in.Status.Id
}

func (in *SapNfsVolume) GetPVLabels() map[string]string {
	result := make(map[string]string, 10)
	if in.Spec.PersistentVolume != nil && in.Spec.PersistentVolume.Labels != nil {
		for k, v := range in.Spec.PersistentVolume.Labels {
			result[k] = v
		}
	}
	result[LabelNfsVolName] = in.Name
	result[LabelNfsVolNS] = in.Namespace
	result[LabelCloudManaged] = "true"

	return result
}

func (in *SapNfsVolume) GetPVAnnotations() map[string]string {
	if in.Spec.PersistentVolume == nil {
		return nil
	}
	result := make(map[string]string, len(in.Spec.PersistentVolume.Annotations))
	for k, v := range in.Spec.PersistentVolume.Annotations {
		result[k] = v
	}
	return result
}

func (in *SapNfsVolume) GetPVCName() string {
	if in.Spec.PersistentVolumeClaim != nil && in.Spec.PersistentVolumeClaim.Name != "" {
		return in.Spec.PersistentVolumeClaim.Name
	}
	return in.Name
}

func (in *SapNfsVolume) GetPVCLabels() map[string]string {
	result := make(map[string]string, 10)
	if in.Spec.PersistentVolumeClaim != nil && in.Spec.PersistentVolumeClaim.Labels != nil {
		for k, v := range in.Spec.PersistentVolumeClaim.Labels {
			result[k] = v
		}
	}
	result[LabelNfsVolName] = in.Name
	result[LabelNfsVolNS] = in.Namespace
	result[LabelCloudManaged] = "true"

	return result
}

func (in *SapNfsVolume) GetPVCAnnotations() map[string]string {
	if in.Spec.PersistentVolumeClaim == nil {
		return nil
	}
	result := make(map[string]string, len(in.Spec.PersistentVolumeClaim.Annotations))
	for k, v := range in.Spec.PersistentVolumeClaim.Annotations {
		result[k] = v
	}
	return result
}

func (in *SapNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *SapNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *SapNfsVolume) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfs
}

func (in *SapNfsVolume) SpecificToProviders() []string {
	return []string{"openstack"}
}

func (in *SapNfsVolume) State() string {
	return in.Status.State
}

func (in *SapNfsVolume) SetState(v string) {
	in.Status.State = v
}

func (in *SapNfsVolume) CloneForPatchStatus() client.Object {
	result := &SapNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SapNfsVolume",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	if result.Status.Conditions == nil {
		result.Status.Conditions = []metav1.Condition{}
	}
	return result
}

func (in *SapNfsVolume) DeriveStateFromConditions() (changed bool) {
	oldState := in.Status.State
	if meta.FindStatusCondition(in.Status.Conditions, ConditionTypeReady) != nil {
		in.Status.State = StateReady
	}
	if meta.FindStatusCondition(in.Status.Conditions, ConditionTypeError) != nil {
		in.Status.State = StateError
	}
	return in.Status.State != oldState
}

// +kubebuilder:object:root=true

// SapNfsVolumeList contains a list of SapNfsVolume
type SapNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SapNfsVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SapNfsVolume{}, &SapNfsVolumeList{})
}
