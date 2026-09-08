package wafpolicy

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// New returns an Action that will provision and deprovision WAF resources in AWS.
// This is the entry point for AWS-specific WAF policy management.
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)

		return composed.ComposeActions(
			"awsWafPolicy",
			createAwsClient,
			loadWebAcl,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"awsWafPolicy-create",
					createWebAcl,
					checkUpdateNeeded,
					updateWebAcl,
					updateStatus,
				),
				composed.ComposeActions(
					"awsWafPolicy-delete",
					deleteWebAcl,
				),
			),
		)(ctx, state)
	}
}
