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
	SubscriptionLabelSecretBindingName = "cloud-manager.kyma-project.io/subscription-binding-name"
	SubscriptionLabelProvider          = "cloud-manager.kyma-project-io/provider"
	SubscriptionLabel                  = "cloud-manager.kyma-project.io/subscription"
)

// SubscriptionSpec defines the desired state of Subscription.
type SubscriptionSpec struct {
	// SecretBindingName specified the SecretBindingName in the Garden
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="SecretBindingName is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="SecretBindingName is required."
	SecretBindingName string `json:"secretBindingName"`
}

// SubscriptionStatus defines the observed state of Subscription.
type SubscriptionStatus struct {
	// +optional
	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`

	Provider ProviderType `json:"provider"`

	SubscriptionInfo *SubscriptionInfo `json:"subscriptionInfo,omitempty"`
}

// SubscriptionInfo specifies subscription info specific for different providers
//
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type SubscriptionInfo struct {
	Gcp       *SubscriptionInfoGcp       `json:"gcp,omitempty"`
	Azure     *SubscriptionInfoAzure     `json:"azure,omitempty"`
	Aws       *SubscriptionInfoAws       `json:"aws,omitempty"`
	OpenStack *SubscriptionInfoOpenStack `json:"openStack,omitempty"`
}

type SubscriptionInfoGcp struct {
	Project string `json:"project"`
}

type SubscriptionInfoAzure struct {
	TenantId       string `json:"tenantId"`
	SubscriptionId string `json:"subscriptionId"`
}

type SubscriptionInfoAws struct {
	Account string `json:"account"`
}

type SubscriptionInfoOpenStack struct {
	DomainName string `json:"domainName"`
	TenantName string `json:"tenantName"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Subscription is the Schema for the subscriptions API.
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec,omitempty"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

func (in *Subscription) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Subscription) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Subscription) CloneForPatchStatus() client.Object {
	result := &Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Scope",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	if result.Status.Conditions == nil {
		result.Status.Conditions = []metav1.Condition{}
	}
	return result
}

// +kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription.
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}
