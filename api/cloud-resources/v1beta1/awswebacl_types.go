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

// AwsWebAclDefaultAction defines the action when no rules match
// Exactly one of Allow or Block must be set
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclDefaultAction struct {
	// Allow action - permit the request
	// +optional
	Allow *AwsWebAclAllowAction `json:"allow,omitempty"`

	// Block action - block the request
	// +optional
	Block *AwsWebAclBlockAction `json:"block,omitempty"`
}

type AwsWebAclAllowAction struct {
	// CustomRequestHandling - Insert custom headers into allowed requests
	// +optional
	CustomRequestHandling *AwsWebAclCustomRequestHandling `json:"customRequestHandling,omitempty"`
}

type AwsWebAclBlockAction struct {
	// CustomResponse - Send custom response for blocked requests
	// +optional
	CustomResponse *AwsWebAclCustomResponse `json:"customResponse,omitempty"`
}

type AwsWebAclCustomRequestHandling struct {
	// InsertHeaders - Headers to insert into the request
	// +kubebuilder:validation:MaxItems=100
	InsertHeaders []AwsWebAclCustomHTTPHeader `json:"insertHeaders"`
}

type AwsWebAclCustomResponse struct {
	// CustomResponseBodyKey - Reference to custom response body in spec.customResponseBodies
	// +optional
	CustomResponseBodyKey string `json:"customResponseBodyKey,omitempty"`

	// ResponseCode - HTTP status code to return (200-599)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=200
	// +kubebuilder:validation:Maximum=599
	ResponseCode int32 `json:"responseCode"`

	// ResponseHeaders - Custom headers to include in response
	// +optional
	// +kubebuilder:validation:MaxItems=100
	ResponseHeaders []AwsWebAclCustomHTTPHeader `json:"responseHeaders,omitempty"`
}

type AwsWebAclCustomHTTPHeader struct {
	// Name - Header name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Value - Header value
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// AwsWebAclSpec defines the desired state of AwsWebAcl
type AwsWebAclSpec struct {
	// DefaultAction specifies what to do when no rules match
	// +kubebuilder:validation:Required
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

	// CustomResponseBodies - Custom response content for block actions
	// +optional
	CustomResponseBodies map[string]AwsWebAclCustomResponseBody `json:"customResponseBodies,omitempty"`

	// TokenDomains - Domains for cross-site CAPTCHA/Challenge token validation
	// +optional
	// +kubebuilder:validation:MaxItems=10
	TokenDomains []string `json:"tokenDomains,omitempty"`

	// CaptchaConfig - Global default CAPTCHA immunity time
	// +optional
	CaptchaConfig *AwsWebAclCaptchaConfig `json:"captchaConfig,omitempty"`

	// ChallengeConfig - Global default Challenge immunity time
	// +optional
	ChallengeConfig *AwsWebAclChallengeConfig `json:"challengeConfig,omitempty"`
}

// AwsWebAclRule defines a single rule in the WebACL
// +kubebuilder:validation:XValidation:rule="size(self.statements) == 1",message="Rule must have exactly 1 statement (ManagedRuleGroup only)"
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

	// OverrideAction for managed rule groups (defaults to None if not specified)
	// +optional
	OverrideAction *AwsWebAclOverrideAction `json:"overrideAction,omitempty"`

	// Statements - Single ManagedRuleGroup statement
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	Statements []AwsWebAclStatement `json:"statements"`

	// VisibilityConfig for rule-specific metrics
	// +optional
	VisibilityConfig *AwsWebAclVisibilityConfig `json:"visibilityConfig,omitempty"`
}

// AwsWebAclRuleAction represents the action to take when a rule matches
// Exactly one action type must be set
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclRuleAction struct {
	// Allow - Permit the request
	// +optional
	Allow *AwsWebAclAllowAction `json:"allow,omitempty"`

	// Block - Block the request
	// +optional
	Block *AwsWebAclBlockAction `json:"block,omitempty"`

	// Count - Count the request but don't take action
	// +optional
	Count *AwsWebAclCountAction `json:"count,omitempty"`

	// Captcha - Require CAPTCHA challenge
	// +optional
	Captcha *AwsWebAclCaptchaAction `json:"captcha,omitempty"`

	// Challenge - Require silent challenge (similar to CAPTCHA but without visual puzzle)
	// +optional
	Challenge *AwsWebAclChallengeAction `json:"challenge,omitempty"`
}

type AwsWebAclCountAction struct {
	// CustomRequestHandling - Insert custom headers
	// +optional
	CustomRequestHandling *AwsWebAclCustomRequestHandling `json:"customRequestHandling,omitempty"`
}

type AwsWebAclCaptchaAction struct {
	// CustomRequestHandling - Insert custom headers
	// +optional
	CustomRequestHandling *AwsWebAclCustomRequestHandling `json:"customRequestHandling,omitempty"`
}

type AwsWebAclChallengeAction struct {
	// CustomRequestHandling - Insert custom headers when challenge token is valid
	// +optional
	CustomRequestHandling *AwsWebAclCustomRequestHandling `json:"customRequestHandling,omitempty"`
}

// AwsWebAclOverrideAction for managed rule groups
// If not specified, defaults to None (use the rule group's default actions)
// At most one action type can be set
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclOverrideAction struct {
	// None - Don't override, use the rule group's action (default if OverrideAction is omitted or empty)
	// +optional
	None *AwsWebAclNoneAction `json:"none,omitempty"`

	// Count - Override all rules to Count
	// +optional
	Count *AwsWebAclCountAction `json:"count,omitempty"`
}

type AwsWebAclNoneAction struct {
	// No fields - empty struct
}

// AwsWebAclStatement - Individual match condition (ManagedRuleGroup only)
type AwsWebAclStatement struct {
	// ManagedRuleGroup - Use AWS-managed rule sets
	// +kubebuilder:validation:Required
	ManagedRuleGroup *AwsWebAclManagedRuleGroupStatement `json:"managedRuleGroup"`
}

// +kubebuilder:validation:XValidation:rule="self.vendorName == 'AWS' && (self.name == 'AWSManagedRulesCommonRuleSet' || self.name == 'AWSManagedRulesKnownBadInputsRuleSet' || self.name == 'AWSManagedRulesSQLiRuleSet' || self.name == 'AWSManagedRulesLinuxRuleSet' || self.name == 'AWSManagedRulesUnixRuleSet')", message="Only free AWS managed rules are supported: AWSManagedRulesCommonRuleSet, AWSManagedRulesKnownBadInputsRuleSet, AWSManagedRulesSQLiRuleSet, AWSManagedRulesLinuxRuleSet, AWSManagedRulesUnixRuleSet. Paid AWS rules and marketplace vendor rules require subscriptions in the service provider's AWS account."
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

	// RuleActionOverrides to override actions for specific rules
	// +optional
	RuleActionOverrides []AwsWebAclRuleActionOverride `json:"ruleActionOverrides,omitempty"`
}

type AwsWebAclRuleActionOverride struct {
	// Name of the rule to override
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ActionToUse to replace the original action
	// +kubebuilder:validation:Required
	ActionToUse *AwsWebAclRuleAction `json:"actionToUse"`
}

type AwsWebAclCustomResponseBody struct {
	// ContentType - Response content type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=TEXT_PLAIN;TEXT_HTML;APPLICATION_JSON
	ContentType string `json:"contentType"`

	// Content - Response body content (max 4,096 bytes per AWS quota)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=4096
	Content string `json:"content"`
}

type AwsWebAclCaptchaConfig struct {
	// ImmunityTime - Seconds a client is exempt after solving CAPTCHA (60-259200)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=259200
	ImmunityTime int64 `json:"immunityTime"`
}

type AwsWebAclChallengeConfig struct {
	// ImmunityTime - Seconds a client is exempt after passing Challenge (60-259200)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=259200
	ImmunityTime int64 `json:"immunityTime"`
}

type AwsWebAclExcludedRule struct {
	// Name of the rule to exclude from the managed rule group
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type AwsWebAclTextTransformation struct {
	// Priority determines the order of transformations (lower = first)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Priority int32 `json:"priority"`

	// Type of transformation
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=NONE;COMPRESS_WHITE_SPACE;HTML_ENTITY_DECODE;LOWERCASE;CMD_LINE;URL_DECODE;BASE64_DECODE;HEX_DECODE;MD5;REPLACE_COMMENTS;ESCAPE_SEQ_DECODE;SQL_HEX_DECODE;CSS_DECODE;JS_DECODE;NORMALIZE_PATH;NORMALIZE_PATH_WIN;REMOVE_NULLS;REPLACE_NULLS;BASE64_DECODE_EXT;URL_DECODE_UNI;UTF8_TO_UNICODE
	Type string `json:"type"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclFieldToMatch struct {
	// UriPath matches the URI path
	// +optional
	UriPath *AwsWebAclUriPath `json:"uriPath,omitempty"`

	// QueryString matches the query string
	// +optional
	QueryString *AwsWebAclQueryString `json:"queryString,omitempty"`

	// Method matches the HTTP method
	// +optional
	Method *AwsWebAclMethod `json:"method,omitempty"`

	// SingleHeader matches a specific header by name
	// +optional
	SingleHeader *AwsWebAclSingleHeader `json:"singleHeader,omitempty"`

	// Body matches the request body
	// +optional
	Body *AwsWebAclBody `json:"body,omitempty"`

	// AllQueryArguments matches all query arguments
	// +optional
	AllQueryArguments *AwsWebAclAllQueryArguments `json:"allQueryArguments,omitempty"`

	// SingleQueryArgument matches a specific query argument by name
	// +optional
	SingleQueryArgument *AwsWebAclSingleQueryArgument `json:"singleQueryArgument,omitempty"`

	// JsonBody matches and parses JSON request body
	// +optional
	JsonBody *AwsWebAclJsonBody `json:"jsonBody,omitempty"`

	// Cookies matches HTTP cookies
	// +optional
	Cookies *AwsWebAclCookies `json:"cookies,omitempty"`

	// Headers matches multiple HTTP headers
	// +optional
	Headers *AwsWebAclHeaders `json:"headers,omitempty"`
}

// AwsWebAclUriPath matches the URI path component of the request
type AwsWebAclUriPath struct {
}

// AwsWebAclQueryString matches the query string component of the request
type AwsWebAclQueryString struct {
}

// AwsWebAclMethod matches the HTTP method
type AwsWebAclMethod struct {
}

// AwsWebAclSingleHeader matches a specific HTTP header by name
type AwsWebAclSingleHeader struct {
	// Name - The name of the header to inspect
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Name string `json:"name"`
}

// AwsWebAclBody matches the request body
type AwsWebAclBody struct {
	// OversizeHandling behavior (CONTINUE/MATCH/NO_MATCH)
	// +optional
	// +kubebuilder:validation:Enum=CONTINUE;MATCH;NO_MATCH
	// +kubebuilder:default=CONTINUE
	OversizeHandling string `json:"oversizeHandling,omitempty"`
}

// AwsWebAclAllQueryArguments matches all query arguments
type AwsWebAclAllQueryArguments struct {
}

// AwsWebAclSingleQueryArgument matches a specific query argument by name
type AwsWebAclSingleQueryArgument struct {
	// Name - The name of the query argument to inspect (max 30 chars, case-insensitive)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=30
	Name string `json:"name"`
}

// AwsWebAclJsonBody matches and parses JSON request body
type AwsWebAclJsonBody struct {
	// MatchPattern - Which JSON elements to inspect
	// +kubebuilder:validation:Required
	MatchPattern AwsWebAclJsonMatchPattern `json:"matchPattern"`

	// MatchScope - Whether to match against keys, values, or all elements
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ALL;KEY;VALUE
	MatchScope string `json:"matchScope"`

	// InvalidFallbackBehavior - What to do if JSON parsing fails
	// MATCH: Treat as matching the rule
	// NO_MATCH: Treat as not matching the rule
	// EVALUATE_AS_STRING: Inspect as plain text
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=MATCH;NO_MATCH;EVALUATE_AS_STRING
	InvalidFallbackBehavior string `json:"invalidFallbackBehavior"`

	// OversizeHandling - What to do when body exceeds inspection limit
	// +optional
	// +kubebuilder:validation:Enum=CONTINUE;MATCH;NO_MATCH
	// +kubebuilder:default=CONTINUE
	OversizeHandling string `json:"oversizeHandling,omitempty"`
}

// AwsWebAclJsonMatchPattern specifies which JSON elements to inspect
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclJsonMatchPattern struct {
	// All - Inspect all JSON elements
	// +optional
	All *AwsWebAclAll `json:"all,omitempty"`

	// IncludedPaths - JSON Pointer paths to inspect (RFC 6901)
	// Example: ["/dogs/0/name", "/dogs/1/name"]
	// +optional
	// +kubebuilder:validation:MinItems=1
	IncludedPaths []string `json:"includedPaths,omitempty"`
}

// AwsWebAclAll is a marker type indicating "match all"
type AwsWebAclAll struct {
}

// AwsWebAclCookies matches HTTP cookies
type AwsWebAclCookies struct {
	// MatchPattern - Which cookies to inspect
	// +kubebuilder:validation:Required
	MatchPattern AwsWebAclCookieMatchPattern `json:"matchPattern"`

	// MatchScope - Whether to match against keys, values, or all cookie data
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ALL;KEY;VALUE
	MatchScope string `json:"matchScope"`

	// OversizeHandling - What to do when cookies exceed inspection limit (8 KB)
	// +optional
	// +kubebuilder:validation:Enum=CONTINUE;MATCH;NO_MATCH
	// +kubebuilder:default=CONTINUE
	OversizeHandling string `json:"oversizeHandling,omitempty"`
}

// AwsWebAclCookieMatchPattern specifies which cookies to inspect
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclCookieMatchPattern struct {
	// All - Inspect all cookies
	// +optional
	All *AwsWebAclAll `json:"all,omitempty"`

	// IncludedCookies - Inspect only these cookie keys
	// +optional
	// +kubebuilder:validation:MinItems=1
	IncludedCookies []string `json:"includedCookies,omitempty"`

	// ExcludedCookies - Inspect all cookies EXCEPT these keys
	// +optional
	// +kubebuilder:validation:MinItems=1
	ExcludedCookies []string `json:"excludedCookies,omitempty"`
}

// AwsWebAclHeaders matches multiple HTTP headers
type AwsWebAclHeaders struct {
	// MatchPattern - Which headers to inspect
	// +kubebuilder:validation:Required
	MatchPattern AwsWebAclHeaderMatchPattern `json:"matchPattern"`

	// MatchScope - Whether to match against keys, values, or all header data
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ALL;KEY;VALUE
	MatchScope string `json:"matchScope"`

	// OversizeHandling - What to do when headers exceed inspection limit (8 KB)
	// +optional
	// +kubebuilder:validation:Enum=CONTINUE;MATCH;NO_MATCH
	// +kubebuilder:default=CONTINUE
	OversizeHandling string `json:"oversizeHandling,omitempty"`
}

// AwsWebAclHeaderMatchPattern specifies which headers to inspect
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclHeaderMatchPattern struct {
	// All - Inspect all headers
	// +optional
	All *AwsWebAclAll `json:"all,omitempty"`

	// IncludedHeaders - Inspect only these header keys
	// +optional
	// +kubebuilder:validation:MinItems=1
	IncludedHeaders []string `json:"includedHeaders,omitempty"`

	// ExcludedHeaders - Inspect all headers EXCEPT these keys
	// +optional
	// +kubebuilder:validation:MinItems=1
	ExcludedHeaders []string `json:"excludedHeaders,omitempty"`
}

type AwsWebAclLabel struct {
	// Name - Label string (1-1024 chars, can contain alphanumeric, underscore, hyphen, colon)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^[0-9A-Za-z_\-:]+$`
	Name string `json:"name"`
}

type AwsWebAclVisibilityConfig struct {
	// CloudWatchMetricsEnabled enables CloudWatch metrics
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	CloudWatchMetricsEnabled bool `json:"cloudWatchMetricsEnabled"`

	// MetricName for CloudWatch (must be unique)
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[0-9A-Za-z_-]+$`
	MetricName string `json:"metricName,omitempty"`

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
