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

	// AssociationConfig - Custom request body size limits for protected resources
	// Note: ALB and AppSync are fixed at 8KB and cannot be customized
	// This config only applies to: CLOUDFRONT, API_GATEWAY, COGNITO_USER_POOL, APP_RUNNER_SERVICE, VERIFIED_ACCESS_INSTANCE
	// +optional
	AssociationConfig *AwsWebAclAssociationConfig `json:"associationConfig,omitempty"`
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

	// Action when rule matches (use for regular rules, mutually exclusive with OverrideAction)
	// +optional
	Action *AwsWebAclRuleActionType `json:"action,omitempty"`

	// OverrideAction for managed rule groups (mutually exclusive with Action)
	// +optional
	OverrideAction *AwsWebAclOverrideAction `json:"overrideAction,omitempty"`

	// Statement defines the match condition (exactly one must be set)
	// +kubebuilder:validation:Required
	Statement AwsWebAclRuleStatement `json:"statement"`

	// RuleLabels - Labels to apply to matching requests (max 100)
	// Can be used with LabelMatchStatement in subsequent rules
	// +optional
	// +kubebuilder:validation:MaxItems=100
	RuleLabels []AwsWebAclLabel `json:"ruleLabels,omitempty"`

	// CaptchaConfig - Per-rule CAPTCHA immunity time override (overrides global setting)
	// +optional
	CaptchaConfig *AwsWebAclCaptchaConfig `json:"captchaConfig,omitempty"`

	// ChallengeConfig - Per-rule Challenge immunity time override (overrides global setting)
	// +optional
	ChallengeConfig *AwsWebAclChallengeConfig `json:"challengeConfig,omitempty"`

	// VisibilityConfig for rule-specific metrics
	// +optional
	VisibilityConfig *AwsWebAclVisibilityConfig `json:"visibilityConfig,omitempty"`
}

// AwsWebAclRuleActionType represents the action to take when a rule matches
// Exactly one action type must be set
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclRuleActionType struct {
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
// Exactly one action type must be set
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type AwsWebAclOverrideAction struct {
	// None - Don't override, use the rule group's action
	// +optional
	None *AwsWebAclNoneAction `json:"none,omitempty"`

	// Count - Override all rules to Count
	// +optional
	Count *AwsWebAclCountAction `json:"count,omitempty"`
}

type AwsWebAclNoneAction struct {
	// No fields - empty struct
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
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

	// LabelMatch - Match based on labels added by previous rules
	// +optional
	LabelMatch *AwsWebAclLabelMatchStatement `json:"labelMatch,omitempty"`

	// SizeConstraint - Match based on request component size
	// +optional
	SizeConstraint *AwsWebAclSizeConstraintStatement `json:"sizeConstraint,omitempty"`

	// SqliMatch - Detect SQL injection attacks
	// +optional
	SqliMatch *AwsWebAclSqliMatchStatement `json:"sqliMatch,omitempty"`

	// XssMatch - Detect cross-site scripting attacks
	// +optional
	XssMatch *AwsWebAclXssMatchStatement `json:"xssMatch,omitempty"`
}

type AwsWebAclSqliMatchStatement struct {
	// FieldToMatch - Part of the request to inspect for SQL injection
	// +kubebuilder:validation:Required
	FieldToMatch AwsWebAclFieldToMatch `json:"fieldToMatch"`

	// TextTransformations - Transformations to apply before inspection
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	TextTransformations []AwsWebAclTextTransformation `json:"textTransformations"`

	// SensitivityLevel - Detection sensitivity: LOW (default, fewer false positives) or HIGH (more detections)
	// +optional
	// +kubebuilder:validation:Enum=LOW;HIGH
	// +kubebuilder:default="LOW"
	SensitivityLevel string `json:"sensitivityLevel,omitempty"`
}

type AwsWebAclXssMatchStatement struct {
	// FieldToMatch - Part of the request to inspect for XSS attacks
	// +kubebuilder:validation:Required
	FieldToMatch AwsWebAclFieldToMatch `json:"fieldToMatch"`

	// TextTransformations - Transformations to apply before inspection
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	TextTransformations []AwsWebAclTextTransformation `json:"textTransformations"`
}

type AwsWebAclSizeConstraintStatement struct {
	// ComparisonOperator - Comparison operator to use
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=EQ;NE;LE;LT;GE;GT
	ComparisonOperator string `json:"comparisonOperator"`

	// Size - Size in bytes to compare against (0 to 21,474,836,480)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=21474836480
	Size int64 `json:"size"`

	// FieldToMatch - Part of the request to inspect
	// +kubebuilder:validation:Required
	FieldToMatch AwsWebAclFieldToMatch `json:"fieldToMatch"`

	// TextTransformations - Transformations to apply before size check
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	TextTransformations []AwsWebAclTextTransformation `json:"textTransformations"`
}

type AwsWebAclLabelMatchStatement struct {
	// Key - Label key to match against (e.g., "aws:acl:name" or namespace like "aws:acl")
	// Can include namespace specifications separated by colons
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^[0-9A-Za-z_\-:]+$`
	Key string `json:"key"`

	// Scope - Match scope: "LABEL" (full label) or "NAMESPACE" (namespace prefix)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=LABEL;NAMESPACE
	Scope string `json:"scope"`
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

	// ForwardedIPConfig for inspecting geo location from forwarded IP headers
	// +optional
	ForwardedIPConfig *AwsWebAclForwardedIPConfig `json:"forwardedIPConfig,omitempty"`
}

type AwsWebAclForwardedIPConfig struct {
	// HeaderName to extract client IP from (e.g., "X-Forwarded-For")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._$-]+$`
	HeaderName string `json:"headerName"`

	// FallbackBehavior when header is missing or invalid
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=MATCH;NO_MATCH
	FallbackBehavior string `json:"fallbackBehavior"`
}

type AwsWebAclRateBasedStatement struct {
	// Limit - Max requests per 5 minutes from a single IP
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=2000000000
	Limit int64 `json:"limit"`

	// ForwardedIPConfig for rate limiting based on forwarded IP headers
	// +optional
	ForwardedIPConfig *AwsWebAclForwardedIPConfig `json:"forwardedIPConfig,omitempty"`
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

	// ManagedRuleGroupConfigs for ATP/ACFP/Bot Control configuration
	// +optional
	ManagedRuleGroupConfigs []AwsWebAclManagedRuleGroupConfig `json:"managedRuleGroupConfigs,omitempty"`

	// RuleActionOverrides to override actions for specific rules
	// +optional
	RuleActionOverrides []AwsWebAclRuleActionOverride `json:"ruleActionOverrides,omitempty"`
}

type AwsWebAclManagedRuleGroupConfig struct {
	// LoginPath for ATP/ACFP (e.g., "/login")
	// +optional
	LoginPath string `json:"loginPath,omitempty"`

	// PayloadType for request body parsing
	// +optional
	// +kubebuilder:validation:Enum=JSON;FORM_ENCODED
	PayloadType string `json:"payloadType,omitempty"`

	// UsernameField for extracting username from requests
	// +optional
	UsernameField *AwsWebAclFieldIdentifier `json:"usernameField,omitempty"`

	// PasswordField for extracting password from requests
	// +optional
	PasswordField *AwsWebAclFieldIdentifier `json:"passwordField,omitempty"`
}

type AwsWebAclFieldIdentifier struct {
	// Identifier path (e.g., "/username" for JSON, "username" for form)
	// +kubebuilder:validation:Required
	Identifier string `json:"identifier"`
}

type AwsWebAclRuleActionOverride struct {
	// Name of the rule to override
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ActionToUse to replace the original action
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Allow;Block;Count;Captcha
	ActionToUse AwsWebAclRuleAction `json:"actionToUse"`
}

type AwsWebAclCustomResponseBody struct {
	// ContentType - Response content type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=TEXT_PLAIN;TEXT_HTML;APPLICATION_JSON
	ContentType string `json:"contentType"`

	// Content - Response body content (max 10,240 bytes)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=10240
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

// AwsWebAclAssociationConfig specifies custom request body inspection limits
// Note: ALB and AppSync are fixed at 8KB. This only applies to CloudFront, API Gateway, etc.
type AwsWebAclAssociationConfig struct {
	// RequestBody - Map of resource type to body size limit
	// Valid keys: CLOUDFRONT, API_GATEWAY, COGNITO_USER_POOL, APP_RUNNER_SERVICE, VERIFIED_ACCESS_INSTANCE
	// Note: ALB and AppSync always use 8KB and cannot be configured here
	// +optional
	RequestBody map[string]AwsWebAclRequestBodyConfig `json:"requestBody,omitempty"`
}

// AwsWebAclRequestBodyConfig defines request body inspection limits
type AwsWebAclRequestBodyConfig struct {
	// DefaultSizeInspectionLimit - Maximum body size to inspect
	// Valid values: KB_8, KB_16, KB_32, KB_48, KB_64
	// Default is KB_16 (16384 bytes) for configurable resources
	// ALB/AppSync are always KB_8 (8192 bytes) regardless of this setting
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=KB_8;KB_16;KB_32;KB_48;KB_64
	DefaultSizeInspectionLimit string `json:"defaultSizeInspectionLimit"`
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

	// TextTransformations to apply before inspecting (required by AWS)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	TextTransformations []AwsWebAclTextTransformation `json:"textTransformations"`
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

// AwsWebAclAndStatement - Logical AND operation combining multiple statements
// All nested statements must match for the And statement to match
type AwsWebAclAndStatement struct {
	// Statements to combine with AND logic (min 2)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	Statements []AwsWebAclRuleStatement `json:"statements"`
}

// AwsWebAclOrStatement - Logical OR operation combining multiple statements
// At least one nested statement must match for the Or statement to match
type AwsWebAclOrStatement struct {
	// Statements to combine with OR logic (min 2)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	Statements []AwsWebAclRuleStatement `json:"statements"`
}

// AwsWebAclNotStatement - Logical NOT operation negating a statement
// Matches when the nested statement does NOT match
type AwsWebAclNotStatement struct {
	// Statement to negate
	// +kubebuilder:validation:Required
	Statement AwsWebAclRuleStatement `json:"statement"`
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
