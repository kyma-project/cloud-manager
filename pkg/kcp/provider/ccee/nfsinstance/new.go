package nfsinstance

import (
	"context"
	"fmt"
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
			conditionsInit,
			networkLoad,
			networkStopWhenNotFound,
			shareNetworkLoad,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"cceeNfsInstance-delete",
				),
				composed.ComposeActions(
					"cceeNfsInstance-create",
					networkLoad,
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
