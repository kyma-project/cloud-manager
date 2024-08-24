package nfsinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	cceemeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/meta"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		nfsState := st.(nfsinstancetypes.State)
		state, err := stateFactory.NewState(ctx, nfsState)
		if err != nil {
			err = fmt.Errorf("error creating new ccee nfsinstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"cceeNfsInstance",
			actions.AddFinalizer,
			conditionsInit,
			networkLoad,
			networkStopWhenNotFound,
			subnetLoad,
			subnetStopWhenNoFound,
			shareNetworkLoad,
			shareLoad,
			accessLoad,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"cceeNfsInstance-delete",
					accessRevoke,
					shareDelete,
					shareNetworkDelete,
					actions.PatchRemoveFinalizer,
					composed.StopAndForgetAction,
				),
				composed.ComposeActions(
					"cceeNfsInstance-create",
					accessMismatchCheck,
					shareNetworkCreate,
					shareCreate,
					shareWaitAvailable,
					shareExpandShrink,
					shareUpdateStatusCapacity,
					accessGrant,
					shareExportRead,
					statusReady,
					composed.StopAndForgetAction,
				),
			),
		)(cceemeta.SetCeeDomainProjectRegion(
			ctx,
			nfsState.Scope().Spec.Scope.OpenStack.DomainName,
			nfsState.Scope().Spec.Scope.OpenStack.TenantName,
			nfsState.Scope().Spec.Region,
		), state)
	}
}
