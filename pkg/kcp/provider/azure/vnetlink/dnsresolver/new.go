package dnsresolver

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(dnsResolverStateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := dnsResolverStateFactory.NewState(ctx, st.(focal.State))

		if err != nil {
			return composed.LogErrorAndReturn(err, "Failed to bootstrap VNetLink DNS resolver state", composed.StopAndForget, ctx)
		}

		return composed.ComposeActions(
			"azureVNetLinkDnsResolver",
			initState,
			initRemoteClient,
			statusCreating,
			loadVNetLink,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureDnsResolverVNetLink-delete",
					deleteVNetLink,
					actions.PatchRemoveCommonFinalizer(),
				),
				composed.ComposeActions(
					"azureDnsResolverVNetLink-non-delete",
					actions.AddCommonFinalizer(),
					composed.If(
						predicateRequireDnsRulesetShootTag,
						loadDnsForwardingRuleset,
						waitDnsForwardingRulesetTag,
					),
					createVNetLink,
					waitVNetLinkCompleted,
					updateStatus,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)

	}
}
