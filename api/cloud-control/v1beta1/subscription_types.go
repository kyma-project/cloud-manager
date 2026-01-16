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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SubscriptionLabelBindingName is set by the Subscription reconciler from .spec.details.garden.bindingName
	SubscriptionLabelBindingName = "cloud-manager.kyma-project.io/binding-name"

	// SubscriptionLabelProvider is set by the Subscription reconciler
	SubscriptionLabelProvider = "cloud-manager.kyma-project-io/provider"

	// SubscriptionLabel should be set on other resources by their reconcilers to indicate to which subscription they belong to
	SubscriptionLabel = "cloud-manager.kyma-project.io/subscription"
)

// SubscriptionSpec defines the desired state of Subscription.
// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Subscription spec is immutable"
type SubscriptionSpec struct {
	// +kubebuilder:validation:Required
	Details SubscriptionDetails `json:"details"`
}

// +kubebuilder:validation:XValidation:rule="((has(self.garden) ? 1 : 0) + (has(self.aws) ? 1 : 0) + (has(self.gcp) ? 1 : 0) + (has(self.azure) ? 1 : 0) + (has(self.openstack) ? 1 : 0)) == 1",message="Exactly one of garden, aws, azure, gcp or openstack must be specified"
type SubscriptionDetails struct {
	Garden    *SubscriptionGarden        `json:"garden,omitempty"`
	Aws       *SubscriptionInfoAws       `json:"aws,omitempty"`
	Azure     *SubscriptionInfoAzure     `json:"azure,omitempty"`
	Gcp       *SubscriptionInfoGcp       `json:"gcp,omitempty"`
	Openstack *SubscriptionInfoOpenStack `json:"openstack,omitempty"`
}

type SubscriptionGarden struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf", message="BindingName is immutable"
	BindingName string `json:"bindingName"`
}

// SubscriptionStatus defines the observed state of Subscription.
type SubscriptionStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Provider ProviderType `json:"provider,omitempty"`

	// +optional
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
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="Project is required"
	Project string `json:"project"`
}

type SubscriptionInfoAzure struct {
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="tenantId is required"
	TenantId string `json:"tenantId"`
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="SubscriptionId is required"
	SubscriptionId string `json:"subscriptionId"`
}

type SubscriptionInfoAws struct {
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="account is required"
	Account string `json:"account"`
}

type SubscriptionInfoOpenStack struct {
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="DomainName is required"
	DomainName string `json:"domainName"`
	// +kubebuilder:validation:XValidation:rule=(size(self) > 0), message="TenantName is required"
	TenantName string `json:"tenantName"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Subscription is the Schema for the subscriptions API.
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec,omitempty"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

// interfaces ====================

func (in *Subscription) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Subscription) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *Subscription) SetObservedGeneration(v int64) {
	in.Status.ObservedGeneration = v
}

// functions ===============

func (in *Subscription) SetStatusProcessing() {
	in.Status.State = string(StateProcessing)
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonProcessing,
		Message:            ReasonProcessing,
	})
}

func (in *Subscription) SetStatusInvalidSpec(msg string) {
	in.Status.State = string(StateWarning)
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonInvalidSpec,
		Message:            msg,
	})
}

func (in *Subscription) SetStatusReady() {
	in.Status.State = string(StateReady)
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonReady,
		Message:            ReasonReady,
	})
}

func (in *Subscription) SetStatusInvalidBinding(msg string) {
	in.Status.State = ReasonInvalidBinding
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonInvalidBinding,
		Message:            msg,
	})
}

func (in *Subscription) SetStatusDeleteWhileUsed(msg string) {
	in.Status.State = ReasonDeleteWhileUsed
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeDeleteWhileUsed,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonDeleteWhileUsed,
		Message:            msg,
	})
}

func (in *Subscription) RemoveStatusDeleteWhileUsed() {
	in.Status.State = string(StateDeleting)
	meta.RemoveStatusCondition(&in.Status.Conditions, ConditionTypeDeleteWhileUsed)
}

func (in *Subscription) SetStatusDeleting() {
	in.Status.State = string(StateDeleting)
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonDeleting,
		Message:            ReasonDeleting,
	})
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
