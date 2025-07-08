package azurerwxpv

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

func loadAzureFileShare(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Load Azure File Share")
	volumeHandle := state.ObjAsPV().Spec.CSI.VolumeHandle
	_, _, fileShareName, _, _, err := azurerwxvolumebackupclient.ParsePvVolumeHandle(volumeHandle)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error parsing azure file share name", err, ctx)
	}

	fileShare, err := state.client.GetFileShare(ctx, volumeHandle)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error loading azure file share", err, ctx)
	}

	//Store it in state
	state.fileShareName = fileShareName
	state.fileShare = fileShare
	if state.fileShare != nil {
		logger.Info(fmt.Sprintf("Loaded FileShare : %v", *state.fileShare.ID))
	}
	return nil, ctx
}
