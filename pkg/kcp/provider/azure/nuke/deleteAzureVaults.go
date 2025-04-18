package nuke

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
)

func deleteAzureVaults(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("deleteAzureVaults")
	for _, rks := range state.ProviderResources {
		if rks.Kind == azurenukeclient.AzureRecoveryVault && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureVault)
				exists, err := state.azureClient.HasProtectedItems(ctx, item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error loading protected items for Azure Vault %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
				if exists {
					logger.Info(fmt.Sprintf("Not deleting vault %v as more protection items exists in vault", item.Name))
					continue
				}

				exists, err = state.azureClient.HasProtectionContainers(ctx, item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error loading registered containers for Azure Vault %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
				if exists {
					logger.Info(fmt.Sprintf("Not deleting vault %v as more protection containers exists in vault", item.Name))
					continue
				}

				err = state.azureClient.DeleteVault(ctx, item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error Deleting Azure Vault: %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
