package awswebacl

import (
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func convertDefaultAction(action cloudresourcesv1beta1.AwsWebAclDefaultAction) (*wafv2types.DefaultAction, error) {
	switch action {
	case cloudresourcesv1beta1.AwsWebAclDefaultActionAllow:
		return &wafv2types.DefaultAction{
			Allow: &wafv2types.AllowAction{},
		}, nil
	case cloudresourcesv1beta1.AwsWebAclDefaultActionBlock:
		return &wafv2types.DefaultAction{
			Block: &wafv2types.BlockAction{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown default action: %s", action)
	}
}

func convertRules(rules []cloudresourcesv1beta1.AwsWebAclRule) ([]wafv2types.Rule, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	result := make([]wafv2types.Rule, 0, len(rules))
	for _, rule := range rules {
		wafRule, err := convertRule(rule)
		if err != nil {
			return nil, fmt.Errorf("error converting rule %s: %w", rule.Name, err)
		}
		result = append(result, *wafRule)
	}
	return result, nil
}

func convertRule(rule cloudresourcesv1beta1.AwsWebAclRule) (*wafv2types.Rule, error) {
	statement, err := convertStatement(rule.Statement)
	if err != nil {
		return nil, err
	}

	visibilityConfig := convertVisibilityConfig(rule.VisibilityConfig, rule.Name)

	wafRule := &wafv2types.Rule{
		Name:             ptr.To(rule.Name),
		Priority:         rule.Priority,
		Statement:        statement,
		VisibilityConfig: visibilityConfig,
	}

	// Managed rule groups require OverrideAction instead of Action
	if rule.Statement.ManagedRuleGroup != nil {
		wafRule.OverrideAction = &wafv2types.OverrideAction{
			None: &wafv2types.NoneAction{},
		}
	} else {
		action, err := convertRuleAction(rule.Action)
		if err != nil {
			return nil, err
		}
		wafRule.Action = action
	}

	return wafRule, nil
}

func convertRuleAction(action cloudresourcesv1beta1.AwsWebAclRuleAction) (*wafv2types.RuleAction, error) {
	switch action {
	case cloudresourcesv1beta1.AwsWebAclRuleActionAllow:
		return &wafv2types.RuleAction{
			Allow: &wafv2types.AllowAction{},
		}, nil
	case cloudresourcesv1beta1.AwsWebAclRuleActionBlock:
		return &wafv2types.RuleAction{
			Block: &wafv2types.BlockAction{},
		}, nil
	case cloudresourcesv1beta1.AwsWebAclRuleActionCount:
		return &wafv2types.RuleAction{
			Count: &wafv2types.CountAction{},
		}, nil
	case cloudresourcesv1beta1.AwsWebAclRuleActionCaptcha:
		return &wafv2types.RuleAction{
			Captcha: &wafv2types.CaptchaAction{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown rule action: %s", action)
	}
}

func convertStatement(stmt cloudresourcesv1beta1.AwsWebAclRuleStatement) (*wafv2types.Statement, error) {
	statement := &wafv2types.Statement{}
	count := 0

	if stmt.IPSet != nil {
		statement.IPSetReferenceStatement = convertIPSetStatement(stmt.IPSet)
		count++
	}

	if stmt.GeoMatch != nil {
		statement.GeoMatchStatement = convertGeoMatchStatement(stmt.GeoMatch)
		count++
	}

	if stmt.RateBased != nil {
		statement.RateBasedStatement = convertRateBasedStatement(stmt.RateBased)
		count++
	}

	if stmt.ManagedRuleGroup != nil {
		statement.ManagedRuleGroupStatement = convertManagedRuleGroupStatement(stmt.ManagedRuleGroup)
		count++
	}

	if count == 0 {
		return nil, fmt.Errorf("statement must have exactly one condition set")
	}
	if count > 1 {
		return nil, fmt.Errorf("statement must have exactly one condition set, found %d", count)
	}

	return statement, nil
}

func convertIPSetStatement(ipSet *cloudresourcesv1beta1.AwsWebAclIPSetStatement) *wafv2types.IPSetReferenceStatement {
	// Note: For inline IP sets, we would need to create an IPSet resource first
	// and then reference it here. This is a placeholder.
	// The actual implementation would need to manage IPSet lifecycle.
	return &wafv2types.IPSetReferenceStatement{
		ARN: ptr.To(""), // This would be populated after creating the IPSet
	}
}

func convertGeoMatchStatement(geo *cloudresourcesv1beta1.AwsWebAclGeoMatchStatement) *wafv2types.GeoMatchStatement {
	countryCodes := make([]wafv2types.CountryCode, 0, len(geo.CountryCodes))
	for _, code := range geo.CountryCodes {
		countryCodes = append(countryCodes, wafv2types.CountryCode(code))
	}
	return &wafv2types.GeoMatchStatement{
		CountryCodes: countryCodes,
	}
}

func convertRateBasedStatement(rate *cloudresourcesv1beta1.AwsWebAclRateBasedStatement) *wafv2types.RateBasedStatement {
	return &wafv2types.RateBasedStatement{
		Limit:               ptr.To(rate.Limit),
		AggregateKeyType:    wafv2types.RateBasedStatementAggregateKeyTypeIp,
		EvaluationWindowSec: 300, // 5 minutes
	}
}

func convertManagedRuleGroupStatement(managed *cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement) *wafv2types.ManagedRuleGroupStatement {
	stmt := &wafv2types.ManagedRuleGroupStatement{
		VendorName: ptr.To(managed.VendorName),
		Name:       ptr.To(managed.Name),
	}

	if managed.Version != "" {
		stmt.Version = ptr.To(managed.Version)
	}

	if len(managed.ExcludedRules) > 0 {
		stmt.ExcludedRules = make([]wafv2types.ExcludedRule, 0, len(managed.ExcludedRules))
		for _, excluded := range managed.ExcludedRules {
			stmt.ExcludedRules = append(stmt.ExcludedRules, wafv2types.ExcludedRule{
				Name: ptr.To(excluded.Name),
			})
		}
	}

	return stmt
}

func convertVisibilityConfig(config *cloudresourcesv1beta1.AwsWebAclVisibilityConfig, defaultName string) *wafv2types.VisibilityConfig {
	// If no visibility config provided, use defaults with metrics/sampling disabled
	if config == nil {
		return &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: false,
			MetricName:               ptr.To(defaultName),
			SampledRequestsEnabled:   false,
		}
	}

	return &wafv2types.VisibilityConfig{
		CloudWatchMetricsEnabled: config.CloudWatchMetricsEnabled,
		MetricName:               ptr.To(config.MetricName),
		SampledRequestsEnabled:   config.SampledRequestsEnabled,
	}
}

func convertScope(scope *cloudcontrolv1beta1.Scope) (wafv2types.Scope, error) {
	// For now, always use REGIONAL
	// CloudFront scope would require us-east-1 region
	return wafv2types.ScopeRegional, nil
}

func convertTags(webAcl *cloudresourcesv1beta1.AwsWebAcl) []wafv2types.Tag {
	tags := []wafv2types.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(webAcl.Name),
		},
		{
			Key:   ptr.To("ManagedBy"),
			Value: ptr.To("cloud-manager"),
		},
	}

	// Add labels as tags
	for key, value := range webAcl.Labels {
		tags = append(tags, wafv2types.Tag{
			Key:   ptr.To(key),
			Value: ptr.To(value),
		})
	}

	return tags
}

// extractIdFromArn extracts the WebACL ID from its ARN
// ARN format: arn:aws:wafv2:region:account:scope/webacl/name/id
func extractIdFromArn(arn string) string {
	if arn == "" {
		return ""
	}
	parts := strings.Split(arn, "/")
	if len(parts) >= 4 {
		return parts[len(parts)-1]
	}
	return ""
}

func isNotFoundError(err error) bool {
	return awsmeta.IsNotFound(err)
}
