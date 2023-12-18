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

const (
	ScopeKymaLabel = "operator.kyma-project.io/kyma-name"
)

// ScopeSpec defines the desired state of Scope
type ScopeSpec struct {
	// +kubebuilder:validation:Required
	Kyma string `json:"kyma"`

	// +kubebuilder:validation:Required
	ShootName string `json:"shootName"`

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
}

type GcpScope struct {
	// +kubebuilder:validation:Required
	Project string `json:"project"`

	// +kubebuilder:validation:Required
	VpcNetwork string `json:"vpcNetwork"`
}

type AzureScope struct {
	// +kubebuilder:validation:Required
	TenantId string `json:"tenantId"`

	// +kubebuilder:validation:Required
	SubscriptionId string `json:"subscriptionId"`

	// +kubebuilder:validation:Required
	VpcNetwork string `json:"vpcNetwork"`
}

type AwsScope struct {
	// +kubebuilder:validation:Required
	Foo string `json:"foo"`
}

// ScopeStatus defines the observed state of Scope
type ScopeStatus struct {
	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Scope is the Schema for the scopes API
type Scope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScopeSpec   `json:"spec,omitempty"`
	Status ScopeStatus `json:"status,omitempty"`
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
