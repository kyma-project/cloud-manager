package nuke

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func disableSoftDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("disableSoftDelete")
	for _, rks := range state.ProviderResources {
		if rks.Kind == "AzureRecoveryVault" && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureVault)
				softDelete := item.Properties.SecuritySettings.SoftDeleteSettings.SoftDeleteState

				if *softDelete != armrecoveryservices.SoftDeleteStateAlwaysON && *softDelete != armrecoveryservices.SoftDeleteStateEnabled {
					continue
				}

				err := state.azureClient.DisableSoftDelete(ctx, item.Vault)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Error Disabling SoftDelete on Azure Vault: %s", obj.GetId()))
				}
			}
		}

	}
	return nil, nil
}
