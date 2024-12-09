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
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:default:=external
// +kubebuilder:validation:Enum=external;kyma;cloud-resources
type NetworkType string

const (
	NetworkTypeExternal       NetworkType = "external"
	NetworkTypeKyma           NetworkType = "kyma"
	NetworkTypeCloudResources NetworkType = "cloud-resources"
)

const NetworkFieldScope = ".spec.scope.name"

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// Scope specifies to which SKR the Network resource belongs to. Managed networks are created in the cloud
	// provider parameters specified in the Scope. If it's a network reference type of Network, it will probably
	// be for some other cloud provider parameters (GCP project, AWS account, Azure tenant/subscription) but
	// the Scope keeps track of the SKR that network was mentioned. The network reference has all parameters
	// required for cloud provider access.
	//
	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Network NetworkInfo `json:"network"`

	// +optional
	// +kubebuilder:default:=external
	Type NetworkType `json:"type,omitempty"`
}

// NetworkInfo can be one of ManagedNetwork or NetworkReference
//
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Network is immutable."
type NetworkInfo struct {
	// Managed described the parameters of the network that will be created and managed by Cloud Manager.
	// The deletion of the Netowork resource will
	// If specified, then Reference must be nil.
	// +optional
	Managed *ManagedNetwork `json:"managed,omitempty"`

	// Reference describes externally managed network that Cloud Manager does not modify but only reads it attributes
	// and uses in other managed resources.
	// +optional
	Reference *NetworkReference `json:"reference,omitempty"`
}

// ManagedNetwork defines parameters for VPC network creation. In Azure and AWS networks must have CIDR address
// space specified, while for GCP it can be empty.
type ManagedNetwork struct {
	// Cidr address range. Not used for GCP. Defaults to default cloud manager cidr "10.250.4.0/22"
	Cidr string `json:"cidr,omitempty"`

	// Location of the network for providers where applicable. If empty the shoot location is used
	Location string `json:"location,omitempty"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type NetworkReference struct {
	// +optional
	Gcp *GcpNetworkReference `json:"gcp,omitempty"`

	// +optional
	Azure *AzureNetworkReference `json:"azure,omitempty"`

	// +optional
	Aws *AwsNetworkReference `json:"aws,omitempty"`

	// +optional
	OpenStack *OpenStackNetworkReference `json:"openstack,omitempty"`
}

func (r *NetworkReference) Equals(other *NetworkReference) bool {
	if r == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(r, other)
}

type GcpNetworkReference struct {
	GcpProject string `json:"gcpProject"`

	NetworkName string `json:"networkName"`
}

type AzureNetworkReference struct {
	TenantId string `json:"tenantId,omitempty"`

	SubscriptionId string `json:"subscriptionId,omitempty"`

	ResourceGroup string `json:"resourceGroup"`

	NetworkName string `json:"networkName"`
}

type AwsNetworkReference struct {
	AwsAccountId string `json:"awsAccountId"`

	Region string `json:"region"`

	VpcId string `json:"vpcId"`

	NetworkName string `json:"networkName"`
}

type OpenStackNetworkReference struct {
	Domain string `json:"domain"`

	Project string `json:"project"`

	NetworkId string `json:"networkId"`

	NetworkName string `json:"networkName"`
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	NetworkType NetworkType `json:"networkType,omitempty"`

	// +optional
	Network *NetworkReference `json:"network,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// Network is the Schema for the networks API
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

func (in *Network) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *Network) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *Network) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Network) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Network) CloneForPatchStatus() client.Object {
	return &Network{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Network",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

func (in *Network) State() string {
	return in.Status.State
}

func (in *Network) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

// FindFirstByType returns first network of the given type.
func (l *NetworkList) FindFirstByType(tp NetworkType) *Network {
	if l == nil {
		return nil
	}
	for _, item := range l.Items {
		if item.Spec.Type == tp {
			return &item
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}
