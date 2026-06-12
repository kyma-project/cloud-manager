package awswebacl

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
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
	statement, err := convertRuleStatement(&rule)
	if err != nil {
		return nil, err
	}

	visibilityConfig := convertVisibilityConfig(rule.VisibilityConfig, rule.Name)

	wafRule := &wafv2types.Rule{
		Name:             &rule.Name,
		Priority:         rule.Priority,
		Statement:        statement,
		VisibilityConfig: visibilityConfig,
	}

	// Since we only support ManagedRuleGroup statements, we only use OverrideAction
	// Default to None if not specified (use the managed rule group's default actions)
	if rule.OverrideAction == nil || (rule.OverrideAction.None == nil && rule.OverrideAction.Count == nil) {
		// Default to None
		wafRule.OverrideAction = &wafv2types.OverrideAction{
			None: &wafv2types.NoneAction{},
		}
	} else {
		overrideAction, err := convertOverrideAction(rule.OverrideAction)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", rule.Name, err)
		}
		wafRule.OverrideAction = overrideAction
	}

	return wafRule, nil
}

func convertRuleActionType(actionType *cloudresourcesv1beta1.AwsWebAclRuleAction) (*wafv2types.RuleAction, error) {
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

	if actionType.Challenge != nil {
		challengeAction := &wafv2types.ChallengeAction{}
		if actionType.Challenge.CustomRequestHandling != nil {
			challengeAction.CustomRequestHandling = convertCustomRequestHandling(actionType.Challenge.CustomRequestHandling)
		}
		result.Challenge = challengeAction
		return result, nil
	}

	return nil, fmt.Errorf("action must have one of allow, block, count, captcha, or challenge set")
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

func convertManagedRuleGroupStatement(managed *cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement) (*wafv2types.ManagedRuleGroupStatement, error) {
	stmt := &wafv2types.ManagedRuleGroupStatement{
		VendorName: &managed.VendorName,
		Name:       &managed.Name,
	}

	if managed.Version != "" {
		stmt.Version = &managed.Version
	}

	if len(managed.ExcludedRules) > 0 {
		stmt.ExcludedRules = make([]wafv2types.ExcludedRule, 0, len(managed.ExcludedRules))
		for _, excluded := range managed.ExcludedRules {
			stmt.ExcludedRules = append(stmt.ExcludedRules, wafv2types.ExcludedRule{
				Name: &excluded.Name,
			})
		}
	}

	if len(managed.RuleActionOverrides) > 0 {
		stmt.RuleActionOverrides = make([]wafv2types.RuleActionOverride, 0, len(managed.RuleActionOverrides))
		for _, override := range managed.RuleActionOverrides {
			action, err := convertRuleActionType(override.ActionToUse)
			if err != nil {
				return nil, fmt.Errorf("error converting rule action override for rule %s: %w", override.Name, err)
			}

			stmt.RuleActionOverrides = append(stmt.RuleActionOverrides, wafv2types.RuleActionOverride{
				Name:        &override.Name,
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
			MetricName:               &defaultName,
			SampledRequestsEnabled:   false,
		}
	}

	// Default metricName to defaultName if not specified
	metricName := config.MetricName
	if metricName == "" {
		metricName = defaultName
	}

	return &wafv2types.VisibilityConfig{
		CloudWatchMetricsEnabled: config.CloudWatchMetricsEnabled,
		MetricName:               &metricName,
		SampledRequestsEnabled:   config.SampledRequestsEnabled,
	}
}

func ScopeRegional() wafv2types.Scope {
	// For now, always use REGIONAL
	return wafv2types.ScopeRegional
}

func convertTags(webAcl *cloudresourcesv1beta1.AwsWebAcl, scope *cloudcontrolv1beta1.Scope) []wafv2types.Tag {
	tags := []wafv2types.Tag{
		{
			Key:   aws.String("Name"),
			Value: &webAcl.Name,
		},
		{
			Key:   aws.String("ManagedBy"),
			Value: aws.String("cloud-manager"),
		},
		{
			Key:   aws.String(common.TagScope),
			Value: &scope.Name,
		},
		{
			Key:   aws.String(common.TagShoot),
			Value: &scope.Spec.ShootName,
		},
	}

	return tags
}

func convertCustomRequestHandling(handling *cloudresourcesv1beta1.AwsWebAclCustomRequestHandling) *wafv2types.CustomRequestHandling {
	if handling == nil || len(handling.InsertHeaders) == 0 {
		return nil
	}

	headers := make([]wafv2types.CustomHTTPHeader, 0, len(handling.InsertHeaders))
	for _, h := range handling.InsertHeaders {
		headers = append(headers, wafv2types.CustomHTTPHeader{
			Name:  &h.Name,
			Value: &h.Value,
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
		ResponseCode: &response.ResponseCode,
	}

	if response.CustomResponseBodyKey != "" {
		result.CustomResponseBodyKey = &response.CustomResponseBodyKey
	}

	if len(response.ResponseHeaders) > 0 {
		headers := make([]wafv2types.CustomHTTPHeader, 0, len(response.ResponseHeaders))
		for _, h := range response.ResponseHeaders {
			headers = append(headers, wafv2types.CustomHTTPHeader{
				Name:  &h.Name,
				Value: &h.Value,
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
			Content:     &body.Content,
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
			ImmunityTime: &config.ImmunityTime,
		},
	}
}

func convertChallengeConfig(config *cloudresourcesv1beta1.AwsWebAclChallengeConfig) *wafv2types.ChallengeConfig {
	if config == nil {
		return nil
	}

	return &wafv2types.ChallengeConfig{
		ImmunityTimeProperty: &wafv2types.ImmunityTimeProperty{
			ImmunityTime: &config.ImmunityTime,
		},
	}
}
