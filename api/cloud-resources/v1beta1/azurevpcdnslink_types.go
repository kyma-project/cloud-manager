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
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureVpcDnsLinkSpec defines the desired state of AzureVpcDnsLink
type AzureVpcDnsLinkSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteLinkName is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 80), message="RemoteLinkName can be up to 80 characters long."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z0-9][a-z0-9-]*[a-z0-9]$') != ''), message="RemoteLinkName must begin with a word character, and it must end with a word character. RemoteLinkName may contain word characters or '-'."
	RemoteLinkName string `json:"remoteLinkName,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemotePrivateDnsZone is immutable."
	RemotePrivateDnsZone string `json:"remotePrivateDnsZone,omitempty"`

	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteTenant is immutable."
	RemoteTenant string `json:"remoteTenant,omitempty"`
}

// AzureVpcDnsLinkStatus defines the observed state of AzureVpcDnsLink
type AzureVpcDnsLinkStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// AzureVpcDnsLink is the Schema for the azurevpcdnslinks API
type AzureVpcDnsLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVpcDnsLinkSpec   `json:"spec,omitempty"`
	Status AzureVpcDnsLinkStatus `json:"status,omitempty"`
}

func (in *AzureVpcDnsLink) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureVpcDnsLink) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureVpcDnsLink) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureVpcDnsLink
}

func (in *AzureVpcDnsLink) SpecificToProviders() []string { return []string{"azure"} }

func (in *AzureVpcDnsLink) State() string { return in.Status.State }

func (in *AzureVpcDnsLink) SetState(v string) { in.Status.State = v }

func (in *AzureVpcDnsLink) Id() string {
	return in.Status.Id
}

func (in *AzureVpcDnsLink) SetId(v string) { in.Status.Id = v }

// +kubebuilder:object:root=true

// AzureVpcDnsLinkList contains a list of AzureVpcDnsLink
type AzureVpcDnsLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVpcDnsLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureVpcDnsLink{}, &AzureVpcDnsLinkList{})
}
