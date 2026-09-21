package wafpolicy

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsWafPolicy()

	// Skip if not created yet
	if webAcl.Status.ProviderId == "" {
		return nil, ctx
	}

	// Check if update is needed
	if !state.updateNeeded {
		return nil, ctx
	}

	logger.Info("Updating AWS WebACL")

	// Parse JSON from spec.data directly into AWS SDK CreateWebACLInput
	var createInput wafv2.CreateWebACLInput
	if err := json.Unmarshal([]byte(webAcl.Spec.Data), &createInput); err != nil {
		logger.Error(err, "Failed to parse spec.data as JSON")
		return composed.LogErrorAndReturn(err, "Error parsing WebACL JSON from spec.data", composed.StopWithRequeue, ctx)
	}

	// Build UpdateWebACLInput from CreateWebACLInput
	input := &wafv2.UpdateWebACLInput{
		Name:                 aws.String(webAcl.Name),
		Id:                   state.awsWebAcl.Id,
		Scope:                ScopeRegional(),
		DefaultAction:        createInput.DefaultAction,
		Rules:                createInput.Rules,
		VisibilityConfig:     createInput.VisibilityConfig,
		CustomResponseBodies: createInput.CustomResponseBodies,
		TokenDomains:         createInput.TokenDomains,
		CaptchaConfig:        createInput.CaptchaConfig,
		ChallengeConfig:      createInput.ChallengeConfig,
		LockToken:            aws.String(state.lockToken),
	}

	if createInput.Description != nil {
		input.Description = createInput.Description
	}

	// Update WebACL
	err := state.awsClient.UpdateWebACL(ctx, input)

	if err != nil {
		logger.Error(err, "Error updating WebACL")
		return composed.LogErrorAndReturn(err, "Error updating AWS WebACL", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL updated successfully, requeueing to reload")

	// Requeue to reload WebACL with fresh state from AWS
	return composed.StopWithRequeue, ctx
}
