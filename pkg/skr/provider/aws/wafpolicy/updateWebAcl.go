package wafpolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if not created yet
	if state.ObjAsWafPolicy().Status.ProviderId == "" {
		return nil, ctx
	}

	// Check if update is needed
	if !state.updateNeeded {
		return nil, ctx
	}

	logger.Info("Updating AWS WebACL")

	createInput := state.parsedInput
	input := &wafv2.UpdateWebACLInput{
		Name:                 aws.String(state.ObjAsWafPolicy().Status.Id),
		Id:                   state.awsWebAcl.Id,
		Scope:                wafv2types.ScopeRegional,
		DefaultAction:        createInput.DefaultAction,
		Rules:                createInput.Rules,
		VisibilityConfig:     createInput.VisibilityConfig,
		CustomResponseBodies: createInput.CustomResponseBodies,
		TokenDomains:         createInput.TokenDomains,
		CaptchaConfig:        createInput.CaptchaConfig,
		ChallengeConfig:      createInput.ChallengeConfig,
		Description:          createInput.Description,
		LockToken:            aws.String(state.lockToken),
	}

	if err := state.awsClient.UpdateWebACL(ctx, input); err != nil {
		logger.Error(err, "Error updating WebACL")
		return composed.LogErrorAndReturn(err, "Error updating AWS WebACL", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL updated successfully, requeueing to reload")
	return composed.StopWithRequeue, ctx
}
