package nuke

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
)

func deleteAzureBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("deleteAzureBackups")
	for _, rks := range state.ProviderResources {
		if rks.Kind == azurenukeclient.AzureFileShareProtection && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				err := state.azureClient.RemoveProtection(ctx, obj.GetId())
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error removing Azure File Backup protection %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
