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
	"github.com/elliotchance/pie/v2"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AlicloudNfsStorageType is the AliCloud NAS file system storage type.
// It mirrors the KCP cloud-control AlicloudStorageType enum; cloud-resources
// types must not import cloud-control, so the enum is duplicated here (the same
// pattern AwsPerformanceMode/AwsThroughputMode follow across both API groups).
// +kubebuilder:validation:Enum=Performance;Capacity;Premium
type AlicloudNfsStorageType string

const (
	AlicloudNfsStorageTypePerformance = AlicloudNfsStorageType("Performance")
	AlicloudNfsStorageTypeCapacity    = AlicloudNfsStorageType("Capacity")
	AlicloudNfsStorageTypePremium     = AlicloudNfsStorageType("Premium")
)

// AlicloudNfsVolumeSpec defines the desired state of AlicloudNfsVolume
type AlicloudNfsVolumeSpec struct {

	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="IpRange is immutable."
	// +optional
	IpRange IpRangeRef `json:"ipRange,omitempty"`
	// Note: omitempty above is required. Without it an unset IpRange serialises as {"name":""},
	// which combined with the self==oldSelf XValidation permanently locks IpRange to empty-string.

	// Capacity sizes the created PersistentVolume/PersistentVolumeClaim. AliCloud
	// NAS Performance/Capacity file systems are elastic (no provisioned size), so
	// this value is not propagated to the KCP NfsInstance.
	// +kubebuilder:validation:Required
	Capacity resource.Quantity `json:"capacity"`

	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="StorageType is immutable."
	// +kubebuilder:default=Performance
	StorageType AlicloudNfsStorageType `json:"storageType,omitempty"`

	PersistentVolume *AlicloudNfsVolumePvSpec `json:"volume,omitempty"`

	PersistentVolumeClaim *AlicloudNfsVolumePvcSpec `json:"volumeClaim,omitempty"`
}

type AlicloudNfsVolumePvSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type AlicloudNfsVolumePvcSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// AlicloudNfsVolumeStatus defines the observed state of AlicloudNfsVolume
type AlicloudNfsVolumeStatus struct {

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

	// +optional
	Capacity resource.Quantity `json:"capacity"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Capacity",type="string",JSONPath=".spec.capacity"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AlicloudNfsVolume is the Schema for the alicloudnfsvolumes API
type AlicloudNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudNfsVolumeSpec   `json:"spec,omitempty"`
	Status AlicloudNfsVolumeStatus `json:"status,omitempty"`
}

func (in *AlicloudNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AlicloudNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AlicloudNfsVolume) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfs
}

func (in *AlicloudNfsVolume) SpecificToProviders() []string {
	return []string{"alicloud"}
}

func (in *AlicloudNfsVolume) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AlicloudNfsVolume) State() string {
	return in.Status.State
}

func (in *AlicloudNfsVolume) SetState(v string) {
	in.Status.State = v
}

func (in *AlicloudNfsVolume) CloneForPatchStatus() client.Object {
	return &AlicloudNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlicloudNfsVolume",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

//+kubebuilder:object:root=true

// AlicloudNfsVolumeList contains a list of AlicloudNfsVolume
type AlicloudNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudNfsVolume `json:"items"`
}

func (l *AlicloudNfsVolumeList) GetItemCount() int {
	return len(l.Items)
}

func (l *AlicloudNfsVolumeList) GetItems() []client.Object {
	return pie.Map(l.Items, func(item AlicloudNfsVolume) client.Object {
		return &item
	})
}

func init() {
	SchemeBuilder.Register(&AlicloudNfsVolume{}, &AlicloudNfsVolumeList{})
}
