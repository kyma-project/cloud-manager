package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		runtimeState := st.(runtimetypes.State)
		cctx, state, err := sf.NewState(ctx, runtimeState)
		if cctx != nil {
			ctx = cctx
		}
		if err != nil {
			return err, ctx
		}

		toggleServiceFlow := composed.ComposeActionsNoName(
			resourceGroupWatcherLoad,
			resourceGroupDataLoad,
			networkWatcherLoad,
			logAnalyticsWorkspaceLoad,
			storageAccountLoad,
			flowLogsLoad,
			composed.IfElse(
				runtimeState.SecurityDataSourceEnabledOnRuntimePredicate,
				// create data sources - enable network flow logs
				composed.ComposeActionsNoName(
					resourceGroupWatcherCreate,
					resourceGroupDataCreate,
					networkWatcherCreate,
					logAnalyticsWorkspaceCreate,
					storageAccountCreate,
					flowLogsCreate,
				),
				// delete data sources - disable network flow logs
				composed.ComposeActionsNoName(
					flowLogsDelete,
					storageAccountDelete,
					logAnalyticsWorkspaceDelete,
					resourceGroupDataDelete,
				),
			),
			// toggle security defender service on/off
			securityPricingLoad,
			securityPlanDefenderCSPM,
			securityPlanServers,
			securityPlanStorage,
			securityPlanResourceManager,
		)

		return composed.ComposeActionsNoName(
			toggleServiceFlow,
		)(ctx, state)
	}
}
