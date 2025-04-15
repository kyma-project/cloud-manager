package nuke

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
)

func deleteAzureContainers(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("UnregisterAzureContainers")

	for _, rks := range state.ProviderResources {
		if rks.Kind == azurenukeclient.AzureStorageContainer && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureContainer)
				if *item.ProtectedItemCount > 0 {
					logger.Info(fmt.Sprintf("Not deleting Storage Container %v as there are %d protected items still present in it", *item.FriendlyName, *item.ProtectedItemCount))
				}

				err := state.azureClient.UnregisterContainer(ctx, obj.GetId())
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error removing Azure Storage Container %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
