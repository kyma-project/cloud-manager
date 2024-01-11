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

// Error reasons
const (
	ReasonInvalidCidr                    = "InvalidCidr"
	ReasonCidrCanNotSplit                = "CidrCanNotSplit"
	ReasonCidrOverlap                    = "CidrOverlap"
	ReasonCidrAssociationFailed          = "CidrAssociationFailed"
	ReasonVpcNotFound                    = "VpcNotFound"
	ReasonShootAndVpcMismatch            = "ShootAndVpcMismatch"
	ReasonFailedExtendingVpcAddressSpace = "FailedExtendingVpcAddressSpace"
	ReasonInvalidIpRangeReference        = "InvalidIpRangeReference"
)

// IpRangeSpec defines the desired state of IpRange
type IpRangeSpec struct {
	// +kubebuilder:validation:Required
	KymaName string `json:"kymaName"`

	// +kubebuilder:validation:Required
	RemoteRef RemoteRef `json:"remoteRef"`

	// +optional
	Scope *ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Cidr string `json:"cidr"`

	// +optional
	Options IpRangeOptions `json:"options,omitempty"`
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

type IpRangeGcp struct {
	// +kubebuilder:default=VPC_PEERING
	Purpose GcpPurpose `json:"purpose,omitempty"`

	// +kubebuilder:default=servicenetworking.googleapis.com
	PsaService string `json:"psaService,omitempty"`
}

type IpRangeAzure struct {
}

type IpRangeAws struct {
}

// IpRangeStatus defines the observed state of IpRange
type IpRangeStatus struct {
	State StatusState `json:"state,omitempty"`

	// +optional
	Ranges []string `json:"ranges,omitempty"`

	// +optional
	VpcId string `json:"vpcId,omitempty"`

	// +optional
	Subnets IpRangeSubnets `json:"subnets,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type IpRangeSubnets []IpRangeSubnet

type IpRangeSubnet struct {
	Id    string `json:"id"`
	Zone  string `json:"zone"`
	Range string `json:"range"`
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

// IpRange is the Schema for the ipranges API
type IpRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpRangeSpec   `json:"spec,omitempty"`
	Status IpRangeStatus `json:"status,omitempty"`
}

func (in *IpRange) KymaName() string {
	return in.Spec.KymaName
}

func (in *IpRange) ScopeRef() *ScopeRef {
	return in.Spec.Scope
}

func (in *IpRange) SetScopeRef(scopeRef *ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *IpRange) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *IpRange) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
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
