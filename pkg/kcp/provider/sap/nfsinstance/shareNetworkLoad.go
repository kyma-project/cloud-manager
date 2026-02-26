package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shareNetworkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	shareNetworkId, _ := state.ObjAsNfsInstance().GetStateData(StateDataShareNetworkId)

	if shareNetworkId == "" {
		arr, err := state.sapClient.ListShareNetworksByNetworkId(ctx, state.network.ID)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing SAP shareNetworks", composed.StopWithRequeue, ctx)
		}
		name := state.ShareNetworkName()
		for _, shareNetwork := range arr {
			if shareNetwork.Name == name {
				state.shareNetwork = &shareNetwork
				break
			}
		}
	} else {
		sn, err := state.sapClient.GetShareNetwork(ctx, shareNetworkId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error getting SAP shareNetwork", composed.StopWithRequeue, ctx)
		}
		state.shareNetwork = sn
	}

	if state.shareNetwork != nil {
		logger = logger.WithValues("sapShareNetworkId", state.shareNetwork.ID)
		ctx = composed.LoggerIntoCtx(ctx, logger)
		//logger.Info("SAP shareNetwork loaded")
	}

	// save shareNetworkId
	if state.shareNetwork != nil && shareNetworkId == "" {
		state.ObjAsNfsInstance().SetStateData(StateDataShareNetworkId, state.shareNetwork.ID)

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating SAP NfsInstance state data with shareNetworkId").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
