package nfsinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shareLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	shareId, _ := state.ObjAsNfsInstance().GetStateData(StateDataShareId)

	if shareId == "" {
		if state.shareNetwork == nil {
			// shareNetwork not loaded so we can not list shares in that shareNetwork
			return nil, nil
		}
		arr, err := state.sapClient.ListSharesInShareNetwork(ctx, state.shareNetwork.ID)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing SAP shares", composed.StopWithRequeue, ctx)
		}
		name := fmt.Sprintf("cm-%s", state.ObjAsNfsInstance().Name)
		for _, share := range arr {
			if share.Name == name {
				state.share = &share
				break
			}
		}
	} else {
		share, err := state.sapClient.GetShare(ctx, shareId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error getting share", composed.StopWithRequeue, ctx)
		}
		state.share = share
	}

	if state.share != nil {
		logger = logger.WithValues("sapShareId", state.share.ID)
		ctx = composed.LoggerIntoCtx(ctx, logger)
		logger.Info("SAP share loaded")
	}

	// save shareId
	if state.share != nil && (shareId == "" || state.ObjAsNfsInstance().Status.Id == "") {
		state.ObjAsNfsInstance().SetStateData(StateDataShareId, state.share.ID)
		state.ObjAsNfsInstance().Status.Id = state.share.ID

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating SAP NfsInstance state data with shareId").
			SuccessErrorNil().
			Run(ctx, state)
	}
	return nil, ctx
}
