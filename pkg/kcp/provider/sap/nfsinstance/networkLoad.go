package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func networkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	networkId, _ := state.ObjAsNfsInstance().GetStateData(StateDataNetworkId)

	if networkId == "" {
		arr, err := state.sapClient.ListInternalNetworksByName(ctx, state.Scope().Spec.Scope.OpenStack.VpcNetwork)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing SAP networks", composed.StopWithRequeue, ctx)
		}
		if len(arr) > 0 {
			state.network = &arr[0]
		}
	} else {
		n, err := state.sapClient.GetNetwork(ctx, networkId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error getting SAP network", composed.StopWithRequeue, ctx)
		}
		state.network = n
	}

	if state.network != nil {
		logger = logger.WithValues("sapNetworkId", state.network.ID)
		ctx = composed.LoggerIntoCtx(ctx, logger)
		//logger.Info("SAP network loaded")
	}

	// save the network id
	if state.network != nil && len(networkId) == 0 {
		state.ObjAsNfsInstance().SetStateData(StateDataNetworkId, state.network.ID)

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating SAP NfsInstance state data with networkId").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
