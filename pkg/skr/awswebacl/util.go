package awswebacl

import (
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/utils/ptr"
)

func convertDefaultAction(action cloudresourcesv1beta1.AwsWebAclDefaultAction) (*wafv2types.DefaultAction, error) {
	result := &wafv2types.DefaultAction{}

	if action.Allow != nil {
		allowAction := &wafv2types.AllowAction{}
		if action.Allow.CustomRequestHandling != nil {
			allowAction.CustomRequestHandling = convertCustomRequestHandling(action.Allow.CustomRequestHandling)
		}
		result.Allow = allowAction
		return result, nil
	}

	if action.Block != nil {
		blockAction := &wafv2types.BlockAction{}
		if action.Block.CustomResponse != nil {
			blockAction.CustomResponse = convertCustomResponse(action.Block.CustomResponse)
		}
		result.Block = blockAction
		return result, nil
	}

	return nil, fmt.Errorf("defaultAction must have either allow or block set")
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

	// Validate: exactly one of Action or OverrideAction must be set
	if rule.Action != nil && rule.OverrideAction != nil {
		return nil, fmt.Errorf("rule %s: cannot set both action and overrideAction", rule.Name)
	}
	if rule.Action == nil && rule.OverrideAction == nil {
		return nil, fmt.Errorf("rule %s: must set either action or overrideAction", rule.Name)
	}

	// Convert Action or OverrideAction
	if rule.Action != nil {
		action, err := convertRuleActionType(rule.Action)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", rule.Name, err)
		}
		wafRule.Action = action
	}

	if rule.OverrideAction != nil {
		overrideAction, err := convertOverrideAction(rule.OverrideAction)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", rule.Name, err)
		}
		wafRule.OverrideAction = overrideAction
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

func convertRuleActionType(actionType *cloudresourcesv1beta1.AwsWebAclRuleActionType) (*wafv2types.RuleAction, error) {
	result := &wafv2types.RuleAction{}

	if actionType.Allow != nil {
		allowAction := &wafv2types.AllowAction{}
		if actionType.Allow.CustomRequestHandling != nil {
			allowAction.CustomRequestHandling = convertCustomRequestHandling(actionType.Allow.CustomRequestHandling)
		}
		result.Allow = allowAction
		return result, nil
	}

	if actionType.Block != nil {
		blockAction := &wafv2types.BlockAction{}
		if actionType.Block.CustomResponse != nil {
			blockAction.CustomResponse = convertCustomResponse(actionType.Block.CustomResponse)
		}
		result.Block = blockAction
		return result, nil
	}

	if actionType.Count != nil {
		countAction := &wafv2types.CountAction{}
		if actionType.Count.CustomRequestHandling != nil {
			countAction.CustomRequestHandling = convertCustomRequestHandling(actionType.Count.CustomRequestHandling)
		}
		result.Count = countAction
		return result, nil
	}

	if actionType.Captcha != nil {
		captchaAction := &wafv2types.CaptchaAction{}
		if actionType.Captcha.CustomRequestHandling != nil {
			captchaAction.CustomRequestHandling = convertCustomRequestHandling(actionType.Captcha.CustomRequestHandling)
		}
		result.Captcha = captchaAction
		return result, nil
	}

	return nil, fmt.Errorf("action must have one of allow, block, count, or captcha set")
}

func convertOverrideAction(overrideAction *cloudresourcesv1beta1.AwsWebAclOverrideAction) (*wafv2types.OverrideAction, error) {
	result := &wafv2types.OverrideAction{}

	if overrideAction.None != nil {
		result.None = &wafv2types.NoneAction{}
		return result, nil
	}

	if overrideAction.Count != nil {
		countAction := &wafv2types.CountAction{}
		if overrideAction.Count.CustomRequestHandling != nil {
			countAction.CustomRequestHandling = convertCustomRequestHandling(overrideAction.Count.CustomRequestHandling)
		}
		result.Count = countAction
		return result, nil
	}

	return nil, fmt.Errorf("overrideAction must have either none or count set")
}

func convertStatement(stmt cloudresourcesv1beta1.AwsWebAclRuleStatement) (*wafv2types.Statement, error) {
	statement := &wafv2types.Statement{}
	count := 0

	if stmt.IPSet != nil {
		statement.IPSetReferenceStatement = convertIPSetStatement(stmt.IPSet)
		count++
	}

	if stmt.GeoMatch != nil {
		geoStmt, err := convertGeoMatchStatement(stmt.GeoMatch)
		if err != nil {
			return nil, err
		}
		statement.GeoMatchStatement = geoStmt
		count++
	}

	if stmt.RateBased != nil {
		rateStmt, err := convertRateBasedStatement(stmt.RateBased)
		if err != nil {
			return nil, err
		}
		statement.RateBasedStatement = rateStmt
		count++
	}

	if stmt.ManagedRuleGroup != nil {
		managedStmt, err := convertManagedRuleGroupStatement(stmt.ManagedRuleGroup)
		if err != nil {
			return nil, err
		}
		statement.ManagedRuleGroupStatement = managedStmt
		count++
	}

	if stmt.ByteMatch != nil {
		byteStmt, err := convertByteMatchStatement(stmt.ByteMatch)
		if err != nil {
			return nil, err
		}
		statement.ByteMatchStatement = byteStmt
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

func convertGeoMatchStatement(geo *cloudresourcesv1beta1.AwsWebAclGeoMatchStatement) (*wafv2types.GeoMatchStatement, error) {
	countryCodes := make([]wafv2types.CountryCode, 0, len(geo.CountryCodes))
	for _, code := range geo.CountryCodes {
		countryCodes = append(countryCodes, wafv2types.CountryCode(code))
	}

	stmt := &wafv2types.GeoMatchStatement{
		CountryCodes: countryCodes,
	}

	if geo.ForwardedIPConfig != nil {
		stmt.ForwardedIPConfig = convertForwardedIPConfig(geo.ForwardedIPConfig)
	}

	return stmt, nil
}

func convertRateBasedStatement(rate *cloudresourcesv1beta1.AwsWebAclRateBasedStatement) (*wafv2types.RateBasedStatement, error) {
	stmt := &wafv2types.RateBasedStatement{
		Limit:               ptr.To(rate.Limit),
		AggregateKeyType:    wafv2types.RateBasedStatementAggregateKeyTypeIp,
		EvaluationWindowSec: 300, // 5 minutes
	}

	if rate.ForwardedIPConfig != nil {
		stmt.ForwardedIPConfig = convertForwardedIPConfig(rate.ForwardedIPConfig)
	}

	return stmt, nil
}

func convertManagedRuleGroupStatement(managed *cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement) (*wafv2types.ManagedRuleGroupStatement, error) {
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

	if len(managed.ManagedRuleGroupConfigs) > 0 {
		stmt.ManagedRuleGroupConfigs = make([]wafv2types.ManagedRuleGroupConfig, 0, len(managed.ManagedRuleGroupConfigs))
		for _, config := range managed.ManagedRuleGroupConfigs {
			wafConfig := wafv2types.ManagedRuleGroupConfig{}

			if config.LoginPath != "" {
				wafConfig.LoginPath = ptr.To(config.LoginPath)
			}

			if config.PayloadType != "" {
				wafConfig.PayloadType = wafv2types.PayloadType(config.PayloadType)
			}

			if config.UsernameField != nil {
				wafConfig.UsernameField = &wafv2types.UsernameField{
					Identifier: ptr.To(config.UsernameField.Identifier),
				}
			}

			if config.PasswordField != nil {
				wafConfig.PasswordField = &wafv2types.PasswordField{
					Identifier: ptr.To(config.PasswordField.Identifier),
				}
			}

			stmt.ManagedRuleGroupConfigs = append(stmt.ManagedRuleGroupConfigs, wafConfig)
		}
	}

	if len(managed.RuleActionOverrides) > 0 {
		stmt.RuleActionOverrides = make([]wafv2types.RuleActionOverride, 0, len(managed.RuleActionOverrides))
		for _, override := range managed.RuleActionOverrides {
			action, err := convertRuleAction(override.ActionToUse)
			if err != nil {
				return nil, fmt.Errorf("error converting rule action override for rule %s: %w", override.Name, err)
			}

			stmt.RuleActionOverrides = append(stmt.RuleActionOverrides, wafv2types.RuleActionOverride{
				Name:        ptr.To(override.Name),
				ActionToUse: action,
			})
		}
	}

	return stmt, nil
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

func ScopeRegional() wafv2types.Scope {
	// For now, always use REGIONAL
	return wafv2types.ScopeRegional
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

func convertByteMatchStatement(byteMatch *cloudresourcesv1beta1.AwsWebAclByteMatchStatement) (*wafv2types.ByteMatchStatement, error) {
	fieldToMatch, err := convertFieldToMatch(byteMatch.FieldToMatch)
	if err != nil {
		return nil, err
	}

	transformations, err := convertTextTransformations(byteMatch.TextTransformations)
	if err != nil {
		return nil, err
	}

	return &wafv2types.ByteMatchStatement{
		SearchString:         []byte(byteMatch.SearchString),
		FieldToMatch:         fieldToMatch,
		PositionalConstraint: wafv2types.PositionalConstraint(byteMatch.PositionalConstraint),
		TextTransformations:  transformations,
	}, nil
}

func convertFieldToMatch(field cloudresourcesv1beta1.AwsWebAclFieldToMatch) (*wafv2types.FieldToMatch, error) {
	result := &wafv2types.FieldToMatch{}

	if field.UriPath {
		result.UriPath = &wafv2types.UriPath{}
		return result, nil
	}

	if field.QueryString {
		result.QueryString = &wafv2types.QueryString{}
		return result, nil
	}

	if field.Method {
		result.Method = &wafv2types.Method{}
		return result, nil
	}

	if field.SingleHeader != "" {
		result.SingleHeader = &wafv2types.SingleHeader{
			Name: ptr.To(field.SingleHeader),
		}
		return result, nil
	}

	if field.Body {
		result.Body = &wafv2types.Body{
			OversizeHandling: wafv2types.OversizeHandlingContinue,
		}
		return result, nil
	}

	return nil, fmt.Errorf("no field to match specified")
}

func convertTextTransformations(transformations []cloudresourcesv1beta1.AwsWebAclTextTransformation) ([]wafv2types.TextTransformation, error) {
	if len(transformations) == 0 {
		return nil, fmt.Errorf("at least one text transformation is required")
	}

	result := make([]wafv2types.TextTransformation, 0, len(transformations))
	for _, t := range transformations {
		result = append(result, wafv2types.TextTransformation{
			Priority: t.Priority,
			Type:     wafv2types.TextTransformationType(t.Type),
		})
	}

	return result, nil
}

func convertForwardedIPConfig(config *cloudresourcesv1beta1.AwsWebAclForwardedIPConfig) *wafv2types.ForwardedIPConfig {
	return &wafv2types.ForwardedIPConfig{
		HeaderName:       ptr.To(config.HeaderName),
		FallbackBehavior: wafv2types.FallbackBehavior(config.FallbackBehavior),
	}
}

func convertCustomRequestHandling(handling *cloudresourcesv1beta1.AwsWebAclCustomRequestHandling) *wafv2types.CustomRequestHandling {
	if handling == nil || len(handling.InsertHeaders) == 0 {
		return nil
	}

	headers := make([]wafv2types.CustomHTTPHeader, 0, len(handling.InsertHeaders))
	for _, h := range handling.InsertHeaders {
		headers = append(headers, wafv2types.CustomHTTPHeader{
			Name:  ptr.To(h.Name),
			Value: ptr.To(h.Value),
		})
	}

	return &wafv2types.CustomRequestHandling{
		InsertHeaders: headers,
	}
}

func convertCustomResponse(response *cloudresourcesv1beta1.AwsWebAclCustomResponse) *wafv2types.CustomResponse {
	if response == nil {
		return nil
	}

	result := &wafv2types.CustomResponse{
		ResponseCode: ptr.To(response.ResponseCode),
	}

	if response.CustomResponseBodyKey != "" {
		result.CustomResponseBodyKey = ptr.To(response.CustomResponseBodyKey)
	}

	if len(response.ResponseHeaders) > 0 {
		headers := make([]wafv2types.CustomHTTPHeader, 0, len(response.ResponseHeaders))
		for _, h := range response.ResponseHeaders {
			headers = append(headers, wafv2types.CustomHTTPHeader{
				Name:  ptr.To(h.Name),
				Value: ptr.To(h.Value),
			})
		}
		result.ResponseHeaders = headers
	}

	return result
}

func convertCustomResponseBodies(bodies map[string]cloudresourcesv1beta1.AwsWebAclCustomResponseBody) map[string]wafv2types.CustomResponseBody {
	if len(bodies) == 0 {
		return nil
	}

	result := make(map[string]wafv2types.CustomResponseBody, len(bodies))
	for key, body := range bodies {
		result[key] = wafv2types.CustomResponseBody{
			ContentType: wafv2types.ResponseContentType(body.ContentType),
			Content:     ptr.To(body.Content),
		}
	}
	return result
}

func convertCaptchaConfig(config *cloudresourcesv1beta1.AwsWebAclCaptchaConfig) *wafv2types.CaptchaConfig {
	if config == nil {
		return nil
	}

	return &wafv2types.CaptchaConfig{
		ImmunityTimeProperty: &wafv2types.ImmunityTimeProperty{
			ImmunityTime: ptr.To(config.ImmunityTime),
		},
	}
}

func convertChallengeConfig(config *cloudresourcesv1beta1.AwsWebAclChallengeConfig) *wafv2types.ChallengeConfig {
	if config == nil {
		return nil
	}

	return &wafv2types.ChallengeConfig{
		ImmunityTimeProperty: &wafv2types.ImmunityTimeProperty{
			ImmunityTime: ptr.To(config.ImmunityTime),
		},
	}
}
