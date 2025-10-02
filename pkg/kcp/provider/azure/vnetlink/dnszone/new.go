package dnszone

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(dnsZoneStateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := dnsZoneStateFactory.NewState(ctx, st.(focal.State))

		if err != nil {
			return composed.LogErrorAndReturn(err, "Failed to bootstrap AzureVNetLink state", composed.StopAndForget, ctx)
		}

		return composed.ComposeActions(
			"azureVNetLink",
			initState,
			initRemoteClient,
			statusInProgress,
			loadVNetLink,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureVNetLink-delete",
					deleteVNetLink,
					actions.PatchRemoveCommonFinalizer(),
				),
				composed.ComposeActions(
					"azureVNetLink-non-delete",
					actions.AddCommonFinalizer(),
					composed.If(
						predicateRequireVNetShootTag,
						loadPrivateDnsZone,
						waitPrivateDnsZoneTag,
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
