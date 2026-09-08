package wafpolicy

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkUpdateNeeded(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsWafPolicy()

	// Skip if not loaded yet
	if state.awsWebAcl == nil {
		return nil, ctx
	}

	// Parse desired spec from JSON
	var desired wafv2.CreateWebACLInput
	if err := json.Unmarshal([]byte(webAcl.Spec.Data), &desired); err != nil {
		return composed.LogErrorAndReturn(err, "Error parsing WebACL JSON from spec.data", composed.StopWithRequeue, ctx)
	}

	// Compare desired vs current state
	current := state.awsWebAcl

	// Check DefaultAction
	if !reflect.DeepEqual(desired.DefaultAction, current.DefaultAction) {
		logger.Info("Update needed: DefaultAction changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check Rules (compare count and deep equality)
	if len(desired.Rules) != len(current.Rules) {
		logger.Info("Update needed: Rules count changed")
		state.updateNeeded = true
		return nil, ctx
	}
	if !reflect.DeepEqual(desired.Rules, current.Rules) {
		logger.Info("Update needed: Rules changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check VisibilityConfig
	if !reflect.DeepEqual(desired.VisibilityConfig, current.VisibilityConfig) {
		logger.Info("Update needed: VisibilityConfig changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check CustomResponseBodies
	if !reflect.DeepEqual(desired.CustomResponseBodies, current.CustomResponseBodies) {
		logger.Info("Update needed: CustomResponseBodies changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check TokenDomains
	if !reflect.DeepEqual(desired.TokenDomains, current.TokenDomains) {
		logger.Info("Update needed: TokenDomains changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check CaptchaConfig
	if !reflect.DeepEqual(desired.CaptchaConfig, current.CaptchaConfig) {
		logger.Info("Update needed: CaptchaConfig changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check ChallengeConfig
	if !reflect.DeepEqual(desired.ChallengeConfig, current.ChallengeConfig) {
		logger.Info("Update needed: ChallengeConfig changed")
		state.updateNeeded = true
		return nil, ctx
	}

	// Check Description
	if !reflect.DeepEqual(desired.Description, current.Description) {
		logger.Info("Update needed: Description changed")
		state.updateNeeded = true
		return nil, ctx
	}

	logger.Info("No update needed, WebACL matches desired state")
	state.updateNeeded = false
	return nil, ctx
}
