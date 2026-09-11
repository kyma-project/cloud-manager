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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WafPolicySpec defines the desired state of WafPolicy
type WafPolicySpec struct {
	// Data contains the provider-specific WAF policy configuration.
	// The structure and format depend on the cloud provider specified in the Scope resource.
	//
	// AWS (supported):
	// AWS WAFv2 WebACL configuration in JSON format, matching the structure of
	// CreateWebACLInput/UpdateWebACLInput, excluding Name and Scope fields (set automatically).
	//
	// Required fields:
	// - DefaultAction: {"Allow": {}} or {"Block": {}}
	// - Rules: array of rule definitions
	// - VisibilityConfig: CloudWatch metrics configuration
	//
	// Example:
	// {
	//   "DefaultAction": {"Allow": {}},
	//   "Rules": [{
	//     "Name": "RateLimitRule",
	//     "Priority": 1,
	//     "Statement": {
	//       "RateBasedStatement": {
	//         "Limit": 2000,
	//         "AggregateKeyType": "IP"
	//       }
	//     },
	//     "Action": {"Block": {}},
	//     "VisibilityConfig": {
	//       "SampledRequestsEnabled": true,
	//       "CloudWatchMetricsEnabled": true,
	//       "MetricName": "RateLimitRule"
	//     }
	//   }],
	//   "VisibilityConfig": {
	//     "SampledRequestsEnabled": true,
	//     "CloudWatchMetricsEnabled": true,
	//     "MetricName": "MyWebACL"
	//   }
	// }
	//
	// See: https://docs.aws.amazon.com/waf/latest/APIReference/API_CreateWebACL.html
	//
	// The configuration is validated by the provider's API during reconciliation.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Data string `json:"data"`
}

// WafPolicyStatus defines the observed state of WafPolicy.
type WafPolicyStatus struct {
	// ProviderId is the provider-specific resource identifier (e.g., AWS ARN, Azure Resource ID)
	// +optional
	ProviderId string `json:"providerId,omitempty"`

	// List of status conditions to indicate the status of a WafPolicy.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// WafPolicy is the Schema for the wafpolicies API
type WafPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of WafPolicy
	// +required
	Spec WafPolicySpec `json:"spec,omitempty"`
	// status defines the observed state of WafPolicy
	// +optional
	Status WafPolicyStatus `json:"status,omitempty"`
}

func (in *WafPolicy) ObservedGeneration() int64 {
	readyCondition := meta.FindStatusCondition(in.Status.Conditions, ConditionTypeReady)
	if readyCondition == nil {
		return 0
	}
	return readyCondition.ObservedGeneration
}

func (in *WafPolicy) SetObservedGeneration(i int64) {
	// ObservedGeneration is managed through the Ready condition
	// This method is kept for interface compatibility but does nothing
}

func (in *WafPolicy) GetStatus() any {
	return &in.Status
}

func (in *WafPolicy) SetStatusProviderError(msg string) {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Generation,
		Reason:             ReasonError,
		Message:            msg,
	})
}

func (in *WafPolicy) SetStatusConfigurationError(msg string) {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonConfigurationError,
		Message:            msg,
	})
}

func (in *WafPolicy) SetStatusFailure(msg string) {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonFailure,
		Message:            msg,
	})
}

func (in *WafPolicy) SetStatusReady() {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonAvailable,
		Message:            ReasonAvailable,
	})
}

func (in *WafPolicy) SetStatusProcessing() {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProcessing,
		Message:            ReasonProcessing,
	})
}

func (in *WafPolicy) SetStatusDeleteWhileUsed(msg string) {
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonDeleteWhileUsed,
		Message:            msg,
	})
}

func (in *WafPolicy) RemoveStatusDeleteWhileUsed() {
	// When DeleteWhileUsed is cleared, set back to Processing to continue deletion
	in.SetStatusProcessing()
}

func (in *WafPolicy) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *WafPolicy) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *WafPolicy) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureWAF
}

func (in *WafPolicy) SpecificToProviders() []string {
	return []string{"aws"}
}

func (in *WafPolicy) State() string {
	readyCondition := meta.FindStatusCondition(in.Status.Conditions, ConditionTypeReady)
	if readyCondition == nil {
		return ""
	}
	return readyCondition.Reason
}

func (in *WafPolicy) SetState(v string) {
	// State is now derived from Ready condition reason
	// This method is kept for interface compatibility but does nothing
}

// +kubebuilder:object:root=true

// WafPolicyList contains a list of WafPolicy
type WafPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WafPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WafPolicy{}, &WafPolicyList{})
}
