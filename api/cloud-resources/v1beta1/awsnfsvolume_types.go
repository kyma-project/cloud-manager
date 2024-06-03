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

// +kubebuilder:validation:Enum=generalPurpose;maxIO
type AwsPerformanceMode string

const (
	AwsPerformanceModeGeneralPurpose = AwsPerformanceMode("generalPurpose")
	AwsPerformanceModeBursting       = AwsPerformanceMode("maxIO")
)

// +kubebuilder:validation:Enum=bursting;elastic
type AwsThroughputMode string

const (
	AwsThroughputModeBursting = AwsThroughputMode("bursting")
	AwsThroughputModeElastic  = AwsThroughputMode("elastic")
)

// AwsNfsVolumeSpec defines the desired state of AwsNfsVolume
type AwsNfsVolumeSpec struct {

	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Capacity resource.Quantity `json:"capacity"`

	// +kubebuilder:default=generalPurpose
	PerformanceMode AwsPerformanceMode `json:"performanceMode,omitempty"`

	// +kubebuilder:default=bursting
	Throughput AwsThroughputMode `json:"throughput,omitempty"`

	PersistentVolume *AwsNfsVolumePvSpec `json:"volume,omitempty"`
}

type AwsNfsVolumePvSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// AwsNfsVolumeStatus defines the observed state of AwsNfsVolume
type AwsNfsVolumeStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	Server string `json:"server,omitempty" json:"server,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Capacity",type="string",JSONPath=".spec.capacity"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// AwsNfsVolume is the Schema for the awsnfsvolumes API
type AwsNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsNfsVolumeSpec   `json:"spec,omitempty"`
	Status AwsNfsVolumeStatus `json:"status,omitempty"`
}

func (in *AwsNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AwsNfsVolume) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfs
}

func (in *AwsNfsVolume) SpecificToProviders() []string {
	return []string{"aws"}
}

func (in *AwsNfsVolume) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AwsNfsVolume) State() string {
	return in.Status.State
}

func (in *AwsNfsVolume) SetState(v string) {
	in.Status.State = v
}

func (in *AwsNfsVolume) CloneForPatchStatus() client.Object {
	return &AwsNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AwsNfsVolume",
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

// AwsNfsVolumeList contains a list of AwsNfsVolume
type AwsNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsNfsVolume `json:"items"`
}

func (l *AwsNfsVolumeList) GetItemCount() int {
	return len(l.Items)
}

func (l *AwsNfsVolumeList) GetItems() []client.Object {
	return pie.Map(l.Items, func(item AwsNfsVolume) client.Object {
		return &item
	})
}

func init() {
	SchemeBuilder.Register(&AwsNfsVolume{}, &AwsNfsVolumeList{})
}
