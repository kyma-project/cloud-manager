package nuke

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteAzureVaults(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("deleteAzureVaults")
	for _, rks := range state.ProviderResources {
		if rks.Kind == "AzureRecoveryVault" && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureVault)
				exists, err := state.azureClient.HasProtectedItems(ctx, item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error loading protected items for Azure Vault %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
				if exists {
					continue
				}

				containerNames, err := state.getContainerNames(item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error getting Azure container names for %s", *item.Name), composed.StopWithRequeue, ctx)
				}
				err = state.azureClient.DeleteVault(ctx, item.Vault, containerNames)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error Deleting Azure Vault: %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
