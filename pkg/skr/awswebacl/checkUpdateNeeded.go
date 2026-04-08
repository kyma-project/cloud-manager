package awswebacl

import (
	"context"
	"reflect"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func checkUpdateNeeded(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if not created yet
	if webAcl.Status.Arn == "" {
		state.updateNeeded = false
		return nil, ctx
	}

	// Skip if AWS WebACL not loaded
	if state.awsWebAcl == nil {
		state.updateNeeded = false
		return nil, ctx
	}

	// Compare defaultAction
	if !compareDefaultAction(state.awsWebAcl.DefaultAction, state.defaultAction) {
		state.updateNeeded = true
		return nil, ctx
	}

	// Compare visibilityConfig
	if !compareVisibilityConfig(state.awsWebAcl.VisibilityConfig, state.visibilityConfig) {
		state.updateNeeded = true
		return nil, ctx
	}

	// Compare rules
	if !compareRules(state.awsWebAcl.Rules, state.rules) {
		state.updateNeeded = true
		return nil, ctx
	}

	// No update needed - stop and forget to skip updateWebAcl
	state.updateNeeded = false
	return nil, ctx
}

func compareDefaultAction(aws, spec *wafv2types.DefaultAction) bool {
	if aws == nil && spec == nil {
		return true
	}
	if aws == nil || spec == nil {
		return false
	}

	// Check Allow
	if (aws.Allow != nil) != (spec.Allow != nil) {
		return false
	}

	// Check Block
	if (aws.Block != nil) != (spec.Block != nil) {
		return false
	}

	return true
}

func compareVisibilityConfig(aws, spec *wafv2types.VisibilityConfig) bool {
	if aws == nil && spec == nil {
		return true
	}
	if aws == nil || spec == nil {
		return false
	}

	if aws.CloudWatchMetricsEnabled != spec.CloudWatchMetricsEnabled {
		return false
	}

	if aws.SampledRequestsEnabled != spec.SampledRequestsEnabled {
		return false
	}

	// MetricName comparison
	if !ptr.Equal(aws.MetricName, spec.MetricName) {
		return false
	}

	return true
}

func compareRules(awsRules, specRules []wafv2types.Rule) bool {
	if len(awsRules) != len(specRules) {
		return false
	}

	// Compare each rule
	for i := range awsRules {
		if !compareRule(&awsRules[i], &specRules[i]) {
			return false
		}
	}

	return true
}

func compareRule(aws, spec *wafv2types.Rule) bool {
	if aws == nil && spec == nil {
		return true
	}
	if aws == nil || spec == nil {
		return false
	}

	// Compare Name
	if !ptr.Equal(aws.Name, spec.Name) {
		return false
	}

	// Compare Priority
	if aws.Priority != spec.Priority {
		return false
	}

	// Compare Action (but only if not a managed rule group)
	if spec.OverrideAction == nil {
		if !compareRuleAction(aws.Action, spec.Action) {
			return false
		}
	}

	// Compare OverrideAction (for managed rule groups)
	if spec.OverrideAction != nil {
		if !compareOverrideAction(aws.OverrideAction, spec.OverrideAction) {
			return false
		}
	}

	// Compare Statement - use deep equal for simplicity
	if !reflect.DeepEqual(aws.Statement, spec.Statement) {
		return false
	}

	// Compare VisibilityConfig
	if !compareVisibilityConfig(aws.VisibilityConfig, spec.VisibilityConfig) {
		return false
	}

	return true
}

func compareRuleAction(aws, spec *wafv2types.RuleAction) bool {
	if aws == nil && spec == nil {
		return true
	}
	if aws == nil || spec == nil {
		return false
	}

	// Check which action type is set
	if (aws.Allow != nil) != (spec.Allow != nil) {
		return false
	}
	if (aws.Block != nil) != (spec.Block != nil) {
		return false
	}
	if (aws.Count != nil) != (spec.Count != nil) {
		return false
	}
	if (aws.Captcha != nil) != (spec.Captcha != nil) {
		return false
	}

	return true
}

func compareOverrideAction(aws, spec *wafv2types.OverrideAction) bool {
	if aws == nil && spec == nil {
		return true
	}
	if aws == nil || spec == nil {
		return false
	}

	// Check None action
	if (aws.None != nil) != (spec.None != nil) {
		return false
	}
	if (aws.Count != nil) != (spec.Count != nil) {
		return false
	}

	return true
}
