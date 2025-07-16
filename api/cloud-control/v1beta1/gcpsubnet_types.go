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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=PRIVATE;
type GcpSubnetPurpose string

const (
	GcpSubnetPurpose_PRIVATE = GcpSubnetPurpose("PRIVATE")
)

const (
	GcpSubnetNetworkField = ".spec.network"
)

// GcpSubnetSpec defines the desired state of GcpSubnet
type GcpSubnetSpec struct {
	// +kubebuilder:validation:Required
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Cidr string `json:"cidr"`

	// +kubebuilder:validation:Required
	Purpose GcpSubnetPurpose `json:"purpose"`

	// Network is a reference to the network where this GcpSubnet belong.
	// If empty then it's implied that it belongs to the Network of the type "kyma" in its Scope.
	// +optional
	Network *klog.ObjectRef `json:"network,omitempty"`
}

// GcpSubnetStatus defines the observed state of GcpSubnet
type GcpSubnetStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	Cidr string `json:"cidr,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="Cidr",type="string",JSONPath=".status.cidr"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpSubnet is the Schema for the gcpsubnets API
type GcpSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpSubnetSpec   `json:"spec,omitempty"`
	Status GcpSubnetStatus `json:"status,omitempty"`
}

func (in *GcpSubnet) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *GcpSubnet) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *GcpSubnet) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpSubnet) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpSubnet) State() string {
	return string(in.Status.State)
}

func (in *GcpSubnet) SetState(v string) {
	in.Status.State = StatusState(v)
}

func (in *GcpSubnet) CloneForPatchStatus() client.Object {
	out := &GcpSubnet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpSubnet",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	return out
}

// +kubebuilder:object:root=true

// GcpSubnetList contains a list of GcpSubnet
type GcpSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpSubnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpSubnet{}, &GcpSubnetList{})
}
