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

// VpcPeeringSpec defines the desired state of VpcPeering
type VpcPeeringSpec struct {
	// +kubebuilder:validation:Required
	Kyma string `json:"kyma"`

	// +kubebuilder:validation:Required
	Scope *ScopeRef `json:"scope"`

	// +optional
	Gcp *GcpVpcPeering `json:"gcp,omitempty"`

	// +optional
	Azure *AzureVpcPeering `json:"azure,omitempty"`

	// +optional
	Aws *AwsVpcPeering `json:"aws,omitempty"`
}

type GcpVpcPeering struct {
	RemoteProject string `json:"remoteProject,omitempty"`
	RemoteVpc     string `json:"remoteVpc,omitempty"`
}

type AzureVpcPeering struct {
	AllowVnetAccess     bool   `json:"allowVnetAccess,omitempty"`
	RemoteVnet          string `json:"remoteVnet,omitempty"`
	RemoteResourceGroup string `json:"remoteResourceGroup,omitempty"`
}

type AwsVpcPeering struct {
	Foo string `json:"foo,omitempty"`
}

// VpcPeeringStatus defines the observed state of VpcPeering
type VpcPeeringStatus struct {
	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VpcPeering is the Schema for the vpcpeerings API
type VpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VpcPeeringSpec   `json:"spec,omitempty"`
	Status VpcPeeringStatus `json:"status,omitempty"`
}

func (in *VpcPeering) KymaName() string {
	return in.Spec.Kyma
}

func (in *VpcPeering) ScopeRef() *ScopeRef {
	return in.Spec.Scope
}

func (in *VpcPeering) SetScopeRef(scopeRef *ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *VpcPeering) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

//+kubebuilder:object:root=true

// VpcPeeringList contains a list of VpcPeering
type VpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VpcPeering{}, &VpcPeeringList{})
}
