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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServedTrue  = "True"
	ServedFalse = "False"

	Finalizer = "cloud-resources.kyma-project.io/deletion-hook"
)

// CloudResourcesSpec defines the desired state of CloudResources
type CloudResourcesSpec struct {
	Aggregations *CloudResourcesAggregation `json:"aggregations,omitempty"`
}

type SourceRef struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name,omitempty"`
}

// ========================================================================

type GcpVpcPeeringInfoList struct {
	Items []*GcpVpcPeeringInfo `json:"items,omitempty"`
}

// ========================================================================

type AzureVpcPeeringInfoList struct {
	Items []*AzureVpcPeeringInfo `json:"items,omitempty"`
}

// ========================================================================

type AwsVpcPeeringInfoList struct {
	Items []*AwsVpcPeeringInfo `json:"items,omitempty"`
}

// ========================================================================

type NfsVolumeInfoList struct {
	Items []*NfsVolumeInfo `json:"items,omitempty"`
}

// ========================================================================

type CloudResourcesAggregation struct {
	GcpVpcPeerings   GcpVpcPeeringInfoList   `json:"gcpVpcPeerings,omitempty"`
	AzureVpcPeerings AzureVpcPeeringInfoList `json:"azureVpcPeerings,omitempty"`
	AwsVpcPeerings   AwsVpcPeeringInfoList   `json:"awsVpcPeerings,omitempty"`
	NfsVolumes       NfsVolumeInfoList       `json:"nfsVolumes,omitempty"`
}

// ========================================================================

type GcpVpcPeeringInfo struct {
	Spec      GcpVpcPeeringSpec `json:"spec"`
	SourceRef SourceRef         `json:"sourceRef"`
}

func (in *GcpVpcPeeringInfo) SetSpec(spec any) {
	in.Spec = spec.(GcpVpcPeeringSpec)
}

func (in *GcpVpcPeeringInfo) GetSourceRef() SourceRef {
	return in.SourceRef
}

// ========================================================================

type AzureVpcPeeringInfo struct {
	Spec      AzureVpcPeeringSpec `json:"spec"`
	SourceRef SourceRef           `json:"sourceRef"`
}

func (in *AzureVpcPeeringInfo) SetSpec(spec any) {
	in.Spec = spec.(AzureVpcPeeringSpec)
}

func (in *AzureVpcPeeringInfo) GetSourceRef() SourceRef {
	return in.SourceRef
}

// ========================================================================

type AwsVpcPeeringInfo struct {
	Spec      AwsVpcPeeringSpec `json:"spec"`
	SourceRef SourceRef         `json:"sourceRef"`
}

func (in *AwsVpcPeeringInfo) SetSpec(spec any) {
	in.Spec = spec.(AwsVpcPeeringSpec)
}

func (in *AwsVpcPeeringInfo) GetSourceRef() SourceRef {
	return in.SourceRef
}

// ========================================================================

type NfsVolumeInfo struct {
	Spec      NfsVolumeSpec `json:"spec"`
	SourceRef SourceRef     `json:"sourceRef"`
}

func (in *NfsVolumeInfo) SetSpec(spec any) {
	in.Spec = spec.(NfsVolumeSpec)
}

func (in *NfsVolumeInfo) GetSourceRef() SourceRef {
	return in.SourceRef
}

// ========================================================================

// CloudResourcesStatus defines the observed state of CloudResources
type CloudResourcesStatus struct {
	State  StatusState `json:"state,omitempty"`
	Served string      `json:"served"`

	// List of status conditions to indicate the status of a CloudResources.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// CloudResources is the Schema for the cloudresources API
type CloudResources struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudResourcesSpec   `json:"spec,omitempty"`
	Status CloudResourcesStatus `json:"status,omitempty"`
}

func (cr *CloudResources) UpdateConditionForReadyState(conditionType ConditionType, reason ConditionReason, conditionStatus metav1.ConditionStatus, message string) {
	cr.Status.State = ReadyState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            message,
	}
	meta.RemoveStatusCondition(&cr.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&cr.Status.Conditions, condition)
}

func (cr *CloudResources) UpdateConditionForErrorState(conditionType ConditionType, reason ConditionReason, conditionStatus metav1.ConditionStatus, error error) {
	cr.Status.State = ErrorState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            error.Error(),
	}
	meta.RemoveStatusCondition(&cr.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&cr.Status.Conditions, condition)
}

//+kubebuilder:object:root=true

// CloudResourcesList contains a list of CloudResources
type CloudResourcesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudResources `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudResources{}, &CloudResourcesList{})
}
