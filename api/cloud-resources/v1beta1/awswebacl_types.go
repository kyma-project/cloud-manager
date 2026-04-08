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
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AwsWebAclDefaultAction string

const (
	AwsWebAclDefaultActionAllow AwsWebAclDefaultAction = "Allow"
	AwsWebAclDefaultActionBlock AwsWebAclDefaultAction = "Block"
)

type AwsWebAclRuleAction string

const (
	AwsWebAclRuleActionAllow   AwsWebAclRuleAction = "Allow"
	AwsWebAclRuleActionBlock   AwsWebAclRuleAction = "Block"
	AwsWebAclRuleActionCount   AwsWebAclRuleAction = "Count"
	AwsWebAclRuleActionCaptcha AwsWebAclRuleAction = "Captcha"
)

// AwsWebAclSpec defines the desired state of AwsWebAcl
type AwsWebAclSpec struct {
	// DefaultAction specifies what to do when no rules match
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Allow;Block
	// +kubebuilder:default=Allow
	DefaultAction AwsWebAclDefaultAction `json:"defaultAction"`

	// Description provides context about the WebACL purpose
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Description string `json:"description,omitempty"`

	// Rules define the filtering logic (evaluated by priority order)
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Rules []AwsWebAclRule `json:"rules,omitempty"`

	// VisibilityConfig defines CloudWatch metrics and request sampling
	// +kubebuilder:validation:Required
	VisibilityConfig *AwsWebAclVisibilityConfig `json:"visibilityConfig"`
}

type AwsWebAclRule struct {
	// Name must be unique within the WebACL
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[0-9A-Za-z_-]+$`
	Name string `json:"name"`

	// Priority determines evaluation order (lower numbers evaluated first, must be unique)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Priority int32 `json:"priority"`

	// Action when rule matches
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Allow;Block;Count;Captcha
	Action AwsWebAclRuleAction `json:"action"`

	// Statement defines the match condition (exactly one must be set)
	// +kubebuilder:validation:Required
	Statement AwsWebAclRuleStatement `json:"statement"`

	// VisibilityConfig for rule-specific metrics
	// +optional
	VisibilityConfig *AwsWebAclVisibilityConfig `json:"visibilityConfig,omitempty"`
}

type AwsWebAclRuleStatement struct {
	// IPSet - Match requests from specific IP addresses/ranges (inline definition)
	// +optional
	IPSet *AwsWebAclIPSetStatement `json:"ipSet,omitempty"`

	// GeoMatch - Match requests from specific countries
	// +optional
	GeoMatch *AwsWebAclGeoMatchStatement `json:"geoMatch,omitempty"`

	// RateBased - Rate limiting per IP
	// +optional
	RateBased *AwsWebAclRateBasedStatement `json:"rateBased,omitempty"`

	// ManagedRuleGroup - Use AWS-managed rule sets
	// +optional
	ManagedRuleGroup *AwsWebAclManagedRuleGroupStatement `json:"managedRuleGroup,omitempty"`

	// ByteMatch - Match specific patterns in requests
	// +optional
	ByteMatch *AwsWebAclByteMatchStatement `json:"byteMatch,omitempty"`
}

type AwsWebAclIPSetStatement struct {
	// IPAddresses in CIDR notation (e.g., "192.0.2.0/24", "2001:db8::/32")
	// Cloud Manager will create/manage the IPSet resource automatically
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10000
	IPAddresses []string `json:"ipAddresses"`
}

type AwsWebAclGeoMatchStatement struct {
	// CountryCodes using ISO 3166-1 alpha-2 codes (e.g., "US", "GB", "DE")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	CountryCodes []string `json:"countryCodes"`
}

type AwsWebAclRateBasedStatement struct {
	// Limit - Max requests per 5 minutes from a single IP
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=2000000000
	Limit int64 `json:"limit"`
}

type AwsWebAclManagedRuleGroupStatement struct {
	// VendorName (typically "AWS" for AWS managed rules)
	// +kubebuilder:validation:Required
	VendorName string `json:"vendorName"`

	// Name of the managed rule group
	// Common AWS managed rules:
	// - AWSManagedRulesCommonRuleSet
	// - AWSManagedRulesKnownBadInputsRuleSet
	// - AWSManagedRulesSQLiRuleSet
	// - AWSManagedRulesLinuxRuleSet
	// - AWSManagedRulesUnixRuleSet
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Version of the rule group (optional, uses latest if not specified)
	// +optional
	Version string `json:"version,omitempty"`

	// ExcludedRules to disable specific rules within the managed group
	// +optional
	ExcludedRules []AwsWebAclExcludedRule `json:"excludedRules,omitempty"`
}

type AwsWebAclExcludedRule struct {
	// Name of the rule to exclude from the managed rule group
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type AwsWebAclByteMatchStatement struct {
	// SearchString to match
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=200
	SearchString string `json:"searchString"`

	// FieldToMatch specifies where to search
	// +kubebuilder:validation:Required
	FieldToMatch AwsWebAclFieldToMatch `json:"fieldToMatch"`

	// PositionalConstraint defines match location
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=EXACTLY;STARTS_WITH;ENDS_WITH;CONTAINS;CONTAINS_WORD
	PositionalConstraint string `json:"positionalConstraint"`
}

type AwsWebAclFieldToMatch struct {
	// UriPath matches the URI path
	// +optional
	UriPath bool `json:"uriPath,omitempty"`

	// QueryString matches the query string
	// +optional
	QueryString bool `json:"queryString,omitempty"`

	// Method matches the HTTP method
	// +optional
	Method bool `json:"method,omitempty"`

	// SingleHeader matches a specific header by name
	// +optional
	SingleHeader string `json:"singleHeader,omitempty"`

	// Body matches the request body
	// +optional
	Body bool `json:"body,omitempty"`
}

type AwsWebAclVisibilityConfig struct {
	// CloudWatchMetricsEnabled enables CloudWatch metrics
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	CloudWatchMetricsEnabled bool `json:"cloudWatchMetricsEnabled"`

	// MetricName for CloudWatch (must be unique)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[0-9A-Za-z_-]+$`
	MetricName string `json:"metricName"`

	// SampledRequestsEnabled enables request sampling in AWS console
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	SampledRequestsEnabled bool `json:"sampledRequestsEnabled"`
}

// AwsWebAclStatus defines the observed state of AwsWebAcl.
type AwsWebAclStatus struct {
	// ARN of the WebACL
	// +optional
	Arn string `json:"arn,omitempty"`

	// Capacity units consumed by the WebACL
	// +optional
	Capacity int64 `json:"capacity,omitempty"`

	// List of status conditions to indicate the status of a AwsWebAcl.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Default Action",type="string",JSONPath=".spec.defaultAction"
// +kubebuilder:printcolumn:name="Capacity",type="integer",JSONPath=".status.capacity"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AwsWebAcl is the Schema for the awswebacls API
type AwsWebAcl struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of AwsWebAcl
	// +required
	Spec AwsWebAclSpec `json:"spec,omitempty"`

	// status defines the observed state of AwsWebAcl
	// +optional
	Status AwsWebAclStatus `json:"status,omitempty"`
}

func (in *AwsWebAcl) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *AwsWebAcl) SetObservedGeneration(i int64) {
	in.Status.ObservedGeneration = i
}

func (in *AwsWebAcl) GetStatus() any {
	return &in.Status
}

func (in *AwsWebAcl) SetStatusProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *AwsWebAcl) SetStatusReady() {
	in.Status.State = StateReady
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonReady,
		Message:            ReasonReady,
	})
}

func (in *AwsWebAcl) SetStatusProcessing() {
	in.Status.State = StateProcessing
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             v1beta1.ReasonProcessing,
		Message:            v1beta1.ReasonProcessing,
	})
}

func (in *AwsWebAcl) SetProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *AwsWebAcl) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AwsWebAcl) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AwsWebAcl) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureWAF
}

func (in *AwsWebAcl) SpecificToProviders() []string { return []string{"aws"} }

func (in *AwsWebAcl) State() string {
	return in.Status.State
}
func (in *AwsWebAcl) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// AwsWebAclList contains a list of AwsWebAcl
type AwsWebAclList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsWebAcl `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsWebAcl{}, &AwsWebAclList{})
}
