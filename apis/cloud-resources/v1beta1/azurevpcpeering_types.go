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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureVpcPeeringSpec defines the desired state of AzureVpcPeering
type AzureVpcPeeringSpec struct {
	AllowVnetAccess     bool   `json:"allowVnetAccess,omitempty"`
	RemoteVnet          string `json:"remoteVnet,omitempty"`
	RemoteResourceGroup string `json:"remoteResourceGroup,omitempty"`
}

// AzureVpcPeeringStatus defines the observed state of AzureVpcPeering
type AzureVpcPeeringStatus struct {
	State State `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// AzureVpcPeering is the Schema for the azurevpcpeerings API
type AzureVpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVpcPeeringSpec   `json:"spec,omitempty"`
	Status AzureVpcPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureVpcPeeringList contains a list of AzureVpcPeering
type AzureVpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVpcPeering `json:"items"`
}

func (peering *AzureVpcPeering) UpdateConditionForReadyState(conditionType ConditionType, reason ConditionReason, conditionStatus metav1.ConditionStatus, message string) {
	peering.Status.State = ReadyState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            message,
	}
	meta.RemoveStatusCondition(&peering.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&peering.Status.Conditions, condition)
}

func (peering *AzureVpcPeering) UpdateConditionForErrorState(conditionType ConditionType, reason ConditionReason, conditionStatus metav1.ConditionStatus, error error) {
	peering.Status.State = ErrorState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            error.Error(),
	}
	meta.RemoveStatusCondition(&peering.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&peering.Status.Conditions, condition)
}

func (me *AzureVpcPeering) GetSourceInfo() SourceRef {
	return SourceRef{
		APIVersion: me.APIVersion,
		Kind:       me.Kind,
		Name:       me.Name,
	}
}

func init() {
	SchemeBuilder.Register(&AzureVpcPeering{}, &AzureVpcPeeringList{})
}
