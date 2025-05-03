package azurerwxpv

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteAzureFileShare(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileShare == nil {
		logger.Info("File Share not found")
		return nil, ctx
	}

	logger.Info("Delete Azure FileShare")
	err := state.client.DeleteFileShare(ctx, state.ObjAsPV().Spec.CSI.VolumeHandle)
	if err != nil {
		//In case of error, reconcile with exponential back off
		return composed.LogErrorAndReturn(err, "error deleting azure file share", err, ctx)
	}

	logger.Info("Waiting for Azure File Share to get Deleted.")
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
