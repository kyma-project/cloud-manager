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
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReasonFailedCreatingVpcPeeringConnection      = "FailedCreatingVpcPeeringConnection"
	ReasonFailedAcceptingVpcPeeringConnection     = "FailedAcceptingVpcPeeringConnection"
	ReasonFailedLoadingRemoteVpcNetwork           = "FailedLoadingRemoteVpcNetwork"
	ReasonFailedLoadingRemoteVpcPeeringConnection = "FailedLoadingRemoteVpcPeeringConnection"
	ReasonFailedCreatingRoutes                    = "FailedCreatingRoutes"
)

const (
	VirtualNetworkPeeringStateConnected    = "Connected"
	VirtualNetworkPeeringStateDisconnected = "Disconnected"
	VirtualNetworkPeeringStateInitiated    = "Initiated"
)

// VpcPeeringSpec defines the desired state of VpcPeering
// +kubebuilder:validation:XValidation:rule=(has(self.vpcPeering) && !has(self.details) || !has(self.vpcPeering) && has(self.details)), message="Only one of details or vpcPeering can be specified."
type VpcPeeringSpec struct {
	// +kubebuilder:validation:Required
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +optional
	VpcPeering *VpcPeeringInfo `json:"vpcPeering"`

	// +optional
	Details *VpcPeeringDetails `json:"details,omitempty"`
}

// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Peering details are immutable."
type VpcPeeringDetails struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self.name != ""), message="Local network name is required."
	LocalNetwork klog.ObjectRef `json:"localNetwork"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self.name != ""), message="Remote network name is required."
	RemoteNetwork klog.ObjectRef `json:"remoteNetwork"`

	PeeringName string `json:"peeringName,omitempty"`

	LocalPeeringName string `json:"localPeeringName,omitempty"`

	ImportCustomRoutes bool `json:"importCustomRoutes,omitempty"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Peering info is immutable."
type VpcPeeringInfo struct {
	// +optional
	Gcp *GcpVpcPeering `json:"gcp,omitempty"`

	// +optional
	Azure *AzureVpcPeering `json:"azure,omitempty"`

	// +optional
	Aws *AwsVpcPeering `json:"aws,omitempty"`
}

type GcpVpcPeering struct {
	RemotePeeringName  string `json:"remotePeeringName,omitempty"`
	RemoteProject      string `json:"remoteProject,omitempty"`
	RemoteVpc          string `json:"remoteVpc,omitempty"`
	ImportCustomRoutes bool   `json:"importCustomRoutes,omitempty"`
}

type AzureVpcPeering struct {
	RemotePeeringName   string `json:"remotePeeringName,omitempty"`
	RemoteVnet          string `json:"remoteVnet,omitempty"`
	RemoteResourceGroup string `json:"remoteResourceGroup,omitempty"`
}

type AwsVpcPeering struct {
	RemoteVpcId     string `json:"remoteVpcId"`
	RemoteRegion    string `json:"remoteRegion,omitempty"`
	RemoteAccountId string `json:"remoteAccountId"`
}

// VpcPeeringStatus defines the observed state of VpcPeering
type VpcPeeringStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// +optional
	VpcId string `json:"vpcId,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	RemoteId string `json:"remoteId,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VpcPeering is the Schema for the vpcpeerings API
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
type VpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VpcPeeringSpec   `json:"spec,omitempty"`
	Status VpcPeeringStatus `json:"status,omitempty"`
}

func (in *VpcPeering) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *VpcPeering) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *VpcPeering) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *VpcPeering) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *VpcPeering) State() string {
	return in.Status.State
}

func (in *VpcPeering) SetState(v string) {
	in.Status.State = v
}

func (in *VpcPeering) CloneForPatchStatus() client.Object {
	return &VpcPeering{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VpcPeering",
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

// VpcPeeringList contains a list of VpcPeering
type VpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VpcPeering{}, &VpcPeeringList{})
}
