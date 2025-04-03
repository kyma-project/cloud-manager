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

// Error reasons
const (
	ReasonCidrCanNotChange               = "CidrCanNotChange"
	ReasonInvalidCidr                    = "InvalidCidr"
	ReasonCidrCanNotSplit                = "CidrCanNotSplit"
	ReasonCidrOverlap                    = "CidrOverlap"
	ReasonCidrAssociationFailed          = "CidrAssociationFailed"
	ReasonCidrAllocationFailed           = "CidrAllocationFailed"
	ReasonVpcNotFound                    = "VpcNotFound"
	ReasonShootAndVpcMismatch            = "ShootAndVpcMismatch"
	ReasonFailedExtendingVpcAddressSpace = "FailedExtendingVpcAddressSpace"
	ReasonInvalidIpRangeReference        = "InvalidIpRangeReference"
)

const (
	IpRangeNetworkField = ".spec.network"
)

// IpRangeSpec defines the desired state of IpRange
type IpRangeSpec struct {
	// +kubebuilder:validation:Required
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +optional
	Cidr string `json:"cidr"`

	// +optional
	Options IpRangeOptions `json:"options,omitempty"`

	// Network is a reference to the network where this IpRange belongs and where it creates subnets.
	// If empty then it's implied that it belongs to the Network of the type "kyma" in its Scope.
	// +optional
	Network *klog.ObjectRef `json:"network,omitempty"`
}

// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type IpRangeOptions struct {
	// +optional
	Gcp *IpRangeGcp `json:"gcp,omitempty"`

	// +optional
	Azure *IpRangeAzure `json:"azure,omitempty"`

	// +optional
	Aws *IpRangeAws `json:"aws,omitempty"`
}

// +kubebuilder:validation:Enum=VPC_PEERING;GCE_ENDPOINT;DNS_RESOLVER;NAT_AUTO;IPSEC_INTERCONNECT;SHARED_LOADBALANCER_VIP;PRIVATE_SERVICE_CONNECT
type GcpPurpose string

const (
	GcpPurposePSA   = GcpPurpose("VPC_PEERING")
	GcpPurposeGCE   = GcpPurpose("GCE_ENDPOINT")
	GcpPurposeDNS   = GcpPurpose("DNS_RESOLVER")
	GcpPurposeNAT   = GcpPurpose("NAT_AUTO")
	GcpPurposeIPSEC = GcpPurpose("IPSEC_INTERCONNECT")
	GcpPurposeVIP   = GcpPurpose("SHARED_LOADBALANCER_VIP")
	GcpPurposePSC   = GcpPurpose("PRIVATE_SERVICE_CONNECT")
)

// +kubebuilder:validation:Enum=GLOBAL_ADDRESS;PRIVATE_SUBNET;
type GcpIpRangeType string

const (
	GcpIpRangeTypeGLOBAL_ADDRESS = GcpIpRangeType("GLOBAL_ADDRESS") // PSA
	GcpIpRangeTypePRIVATE_SUBNET = GcpIpRangeType("PRIVATE_SUBNET")
)

type IpRangeGcp struct {
	// +kubebuilder:default=VPC_PEERING
	Purpose GcpPurpose `json:"purpose,omitempty"`

	// +kubebuilder:default=servicenetworking.googleapis.com
	PsaService string `json:"psaService,omitempty"`

	// +kubebuilder:default=GLOBAL_ADDRESS
	Type GcpIpRangeType `json:"type,omitempty"`
}

type IpRangeAzure struct {
}

type IpRangeAws struct {
}

// IpRangeStatus defines the observed state of IpRange
type IpRangeStatus struct {
	State StatusState `json:"state,omitempty"`

	// +optional
	Cidr string `json:"cidr,omitempty"`

	// +optional
	Ranges []string `json:"ranges,omitempty"`

	// +optional
	VpcId string `json:"vpcId,omitempty"`

	// +optional
	AddressSpaceId string `json:"addressSpaceId,omitempty"`

	// +optional
	Subnets IpRangeSubnets `json:"subnets,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Operation Identifier to track the Hyperscaler Operation
	// +optional
	OpIdentifier string `json:"opIdentifier,omitempty"`

	// Id to track the Hyperscaler IpRange identifier
	// +optional
	Id string `json:"id,omitempty"`
}

type IpRangeSubnets []IpRangeSubnet

type IpRangeSubnet struct {
	Id    string `json:"id"`
	Zone  string `json:"zone"`
	Range string `json:"range"`
}

func (in IpRangeSubnets) Equals(other IpRangeSubnets) bool {
	if len(in) != len(other) {
		return false
	}
	for _, mine := range in {
		found := false
		for _, his := range other {
			if mine == his {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (in IpRangeSubnets) SubnetById(id string) *IpRangeSubnet {
	for _, s := range in {
		if s.Id == id {
			return &s
		}
	}
	return nil
}

func (in IpRangeSubnets) SubnetByZone(zone string) *IpRangeSubnet {
	for _, s := range in {
		if s.Zone == zone {
			return &s
		}
	}
	return nil
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="Network",type="string",JSONPath=".spec.network.name"
// +kubebuilder:printcolumn:name="Cidr",type="string",JSONPath=".status.cidr"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// IpRange is the Schema for the ipranges API
type IpRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpRangeSpec   `json:"spec,omitempty"`
	Status IpRangeStatus `json:"status,omitempty"`
}

func (in *IpRange) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *IpRange) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *IpRange) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *IpRange) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *IpRange) State() string {
	return string(in.Status.State)
}

func (in *IpRange) SetState(v string) {
	in.Status.State = StatusState(v)
}

func (in *IpRange) CloneForPatchStatus() client.Object {
	out := &IpRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IpRange",
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

//+kubebuilder:object:root=true

// IpRangeList contains a list of IpRange
type IpRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpRange `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpRange{}, &IpRangeList{})
}
