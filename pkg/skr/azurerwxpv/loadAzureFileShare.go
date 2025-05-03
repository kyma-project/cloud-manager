package azurerwxpv

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadAzureFileShare(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Load Azure File Share")
	fileShare, err := state.client.GetFileShare(ctx, state.ObjAsPV().Spec.CSI.VolumeHandle)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error loading azure file share", err, ctx)
	}

	//Store it in state
	state.fileShare = fileShare
	return nil, ctx
}
