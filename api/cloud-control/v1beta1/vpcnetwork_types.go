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
	"slices"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=kyma;gardener
type VpcNetworkType string

const (
	VpcNetworkTypeGardener VpcNetworkType = "gardener"
	VpcNetworkTypeKyma     VpcNetworkType = "kyma"
)

// VpcNetworkSpec defines the desired state of VpcNetwork.
type VpcNetworkSpec struct {
	// +optional
	// +kubebuilder:default=kyma
	Type VpcNetworkType `json:"type,omitempty"`

	// +kubebuilder:validation:Required
	Subscription string `json:"subscription"`

	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self.all(s, isCIDR(s))",message="cidrBlocks must be a list of valid CIDRs, for example '10.20.30.40/16'"
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	CidrBlocks []string `json:"cidrBlocks"`
}

// VpcNetworkStatus defines the observed state of VpcNetwork.
type VpcNetworkStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	CidrBlocks []string `json:"cidrBlocks"`

	// +optional
	Identifiers VpcNetworkStatusIdentifiers `json:"identifiers"`
}

type VpcNetworkStatusIdentifiers struct {
	// +optional
	Vpc string `json:"vpc"`

	// +optional
	Router string `json:"router"`

	// +optional
	InternetGateway string `json:"internetGateway"`

	// +optional
	ResourceGroup string `json:"resourceGroup"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Subscription",type="string",JSONPath=".spec.subscription"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// VpcNetwork is the Schema for the vpcnetworks API.
type VpcNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VpcNetworkSpec   `json:"spec,omitempty"`
	Status VpcNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VpcNetworkList contains a list of VpcNetwork.
type VpcNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VpcNetwork `json:"items"`
}

// interface implementations

func (in *VpcNetwork) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *VpcNetwork) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *VpcNetwork) SetObservedGeneration(v int64) {
	in.Status.ObservedGeneration = v
}

// functions

func (in *VpcNetwork) HaveVpcCidrBlocksChanged() bool {
	if len(in.Spec.CidrBlocks) != len(in.Status.CidrBlocks) {
		return true
	}
	for _, x := range in.Spec.CidrBlocks {
		if !slices.Contains(in.Status.CidrBlocks, x) {
			return true
		}
	}

	return false
}

func (in *VpcNetwork) SetStatusProcessing() {
	in.Status.State = ReasonProcessing
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProcessing,
		Message:            ReasonProcessing,
	})
}

func (in *VpcNetwork) SetStatusProvisioned() {
	in.Status.State = ReasonProvisioned
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProvisioned,
		Message:            ReasonProvisioned,
	})
}

func (in *VpcNetwork) SetStatusInvalidCidr(msg string) {
	in.Status.State = ReasonInvalidCidr
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonInvalidCidr,
		Message:            msg,
	})
}

func (in *VpcNetwork) SetStatusOverlappingCidr(msg string) {
	in.Status.State = ReasonCidrOverlap
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonCidrOverlap,
		Message:            msg,
	})
}

func (in *VpcNetwork) SetStatusInvalidDependency(msg string) {
	in.Status.State = ReasonInvalidDependency
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonInvalidDependency,
		Message:            msg,
	})
}

func (in *VpcNetwork) SetStatusProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *VpcNetwork) SetStatusDeleteWhileUsed(msg string) {
	in.Status.State = ReasonDeleteWhileUsed
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeDeleteWhileUsed,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonDeleteWhileUsed,
		Message:            msg,
	})
}

func (in *VpcNetwork) RemoveStatusDeleteWhileUsed() {
	in.Status.State = string(StateDeleting)
	meta.RemoveStatusCondition(&in.Status.Conditions, ConditionTypeDeleteWhileUsed)
}

func (in *VpcNetwork) SetStatusDeleting() {
	in.Status.State = string(StateDeleting)
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonDeleting,
		Message:            ReasonDeleting,
	})
}

func init() {
	SchemeBuilder.Register(&VpcNetwork{}, &VpcNetworkList{})
}
