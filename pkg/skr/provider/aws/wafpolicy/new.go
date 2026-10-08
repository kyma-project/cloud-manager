package wafpolicy

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	wafpolicytypes "github.com/kyma-project/cloud-manager/pkg/skr/wafpolicy/types"
)

// New returns an Action that will provision and deprovision WAF resources in AWS.
// This is the entry point for AWS-specific WAF policy management.
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		wafPolicyState := st.(wafpolicytypes.State)
		state, err := stateFactory.NewState(ctx, wafPolicyState)
		if err != nil {
			err = fmt.Errorf("error creating new aws wafpolicy state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, ctx
		}

		return composed.ComposeActions(
			"awsWafPolicy",
			createAwsClient,
			loadWebAcl,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActionsNoName(
					actions.AddCommonFinalizer(),
					updateId,
					parsePayload,
					createWebAcl,
					checkUpdateNeeded,
					updateWebAcl,
					updateStatus,
				),
				composed.ComposeActions(
					"awsWafPolicy-delete",
					deleteWebAcl,
					actions.RemoveCommonFinalizer(),
				),
			),
		)(ctx, state)
	}
}
