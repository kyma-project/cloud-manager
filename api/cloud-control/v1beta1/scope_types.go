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

// ScopeSpec defines the desired state of Scope
type ScopeSpec struct {
	// +kubebuilder:validation:Required
	KymaName string `json:"kymaName"`

	// +kubebuilder:validation:Required
	ShootName string `json:"shootName"`

	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// +kubebuilder:validation:Required
	Provider ProviderType `json:"provider"`

	// +kubebuilder:validation:Required
	Scope ScopeInfo `json:"scope"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type ScopeInfo struct {
	// +optional
	Gcp *GcpScope `json:"gcp,omitempty"`

	// +optional
	Azure *AzureScope `json:"azure,omitempty"`

	// +optional
	Aws *AwsScope `json:"aws,omitempty"`

	// +optional
	OpenStack *OpenStackScope `json:"openstack,omitempty"`
}

type OpenStackScope struct {
	VpcNetwork string `json:"vpcNetwork"`
	DomainName string `json:"domainName"`
	TenantName string `json:"tenantName"`

	Network OpenStackNetwork `json:"network"`
}

type OpenStackNetwork struct {
	// +optional
	Nodes string `json:"nodes,omitempty"`

	// +optional
	Pods string `json:"pods,omitempty"`

	// +optional
	Services string `json:"services,omitempty"`

	// +optional
	Zones []string `json:"zones,omitempty"`
}

type GcpScope struct {
	// +kubebuilder:validation:Required
	Project string `json:"project"`

	// +kubebuilder:validation:Required
	VpcNetwork string `json:"vpcNetwork"`

	// +optional
	Network GcpNetwork `json:"network"`

	// +optional
	Workers []GcpWorkers `json:"workers"`
}

type GcpWorkers struct {
	// +kubebuilder:validation:Required
	Zones []string `json:"zones"`
}

type GcpNetwork struct {
	// +optional
	Nodes string `json:"nodes,omitempty"`

	// +optional
	Pods string `json:"pods,omitempty"`

	// +optional
	Services string `json:"services,omitempty"`
}

type AzureScope struct {
	// +kubebuilder:validation:Required
	TenantId string `json:"tenantId"`

	// +kubebuilder:validation:Required
	SubscriptionId string `json:"subscriptionId"`

	// +kubebuilder:validation:Required
	VpcNetwork string `json:"vpcNetwork"`

	Network AzureNetwork `json:"network"`
}

type AzureNetwork struct {
	// +optional
	Cidr string `json:"cidr,omitempty"`

	// +optional
	Zones []AzureNetworkZone `json:"zones,omitempty"`

	// +optional
	Nodes string `json:"nodes,omitempty"`

	// +optional
	Pods string `json:"pods,omitempty"`

	// +optional
	Services string `json:"services,omitempty"`
}

type AzureNetworkZone struct {
	Name string `json:"name,omitempty"`
	Cidr string `json:"cidr,omitempty"`
}

type AwsScope struct {
	// +kubebuilder:validation:Required
	VpcNetwork string `json:"vpcNetwork"`

	Network AwsNetwork `json:"network"`

	// +kubebuilder:validation:Required
	AccountId string `json:"accountId"`
}

type AwsNetwork struct {
	// +optional
	Nodes string `json:"nodes,omitempty"`

	// +optional
	Pods string `json:"pods,omitempty"`

	// +optional
	Services string `json:"services,omitempty"`

	VPC   AwsVPC    `json:"vpc"`
	Zones []AwsZone `json:"zones"`
}

type AwsVPC struct {
	Id   string `json:"id,omitempty"`
	CIDR string `json:"cidr,omitempty"`
}

type AwsZone struct {
	Name     string `json:"name"`
	Internal string `json:"internal"`
	Public   string `json:"public"`
	Workers  string `json:"workers"`
}

// ScopeStatus defines the observed state of Scope
type ScopeStatus struct {
	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Operation Identifier to track the ServiceUsage Operation
	// +optional
	GcpOperations []string `json:"gcpOperations,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Scope is the Schema for the scopes API
// +kubebuilder:printcolumn:name="Kyma",type="string",JSONPath=".spec.kymaName"
// +kubebuilder:printcolumn:name="Shoot",type="string",JSONPath=".spec.shootName"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
type Scope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScopeSpec   `json:"spec,omitempty"`
	Status ScopeStatus `json:"status,omitempty"`
}

func (in *Scope) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Scope) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

//+kubebuilder:object:root=true

// ScopeList contains a list of Scope
type ScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scope `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Scope{}, &ScopeList{})
}
