package nuke

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteAzureBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("deleteAzureBackups")
	for _, rks := range state.ProviderResources {
		if rks.Kind == "AzureRwxVolumeBackup" && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureProtectedItem)

				err := state.azureClient.RemoveProtection(ctx, item.ProtectedItemResource)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error removing Azure File Backup protection %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
