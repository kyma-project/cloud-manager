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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReasonFailedLoadingPrivateDnzZone = "FailedLoadingPrivateDnzZone"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureVNetLinkSpec defines the desired state of AzureVNetLink
type AzureVNetLinkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVirtualPrivateLinkName is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 80), message="RemoteVirtualPrivateLinkName can be up to 80 characters long."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z0-9][a-z0-9-]*[a-z0-9]$') != ''), message="RemoteVirtualPrivateLinkName must begin with a word character, and it must end with a word character. RemoteVirtualPrivateLinkName may contain word characters or '-'."
	RemoteVirtualPrivateLinkName string `json:"remoteVirtualPrivateLinkName,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemotePrivateDnsZone is immutable."
	RemotePrivateDnsZone string `json:"remotePrivateDnsZone,omitempty"`

	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteTenant is immutable."
	RemoteTenant string `json:"remoteTenant,omitempty"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`
}

// AzureVNetLinkStatus defines the observed state of AzureVNetLink
type AzureVNetLinkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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

// AzureVNetLink is the Schema for the azurevnetlinks API
type AzureVNetLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVNetLinkSpec   `json:"spec,omitempty"`
	Status AzureVNetLinkStatus `json:"status,omitempty"`
}

func (in *AzureVNetLink) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureVNetLink) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureVNetLink) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *AzureVNetLink) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *AzureVNetLink) CloneForPatchStatus() client.Object {
	return &AzureVNetLink{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureVNetLink",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

// +kubebuilder:object:root=true

// AzureVNetLinkList contains a list of AzureVNetLink
type AzureVNetLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVNetLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureVNetLink{}, &AzureVNetLinkList{})
}
