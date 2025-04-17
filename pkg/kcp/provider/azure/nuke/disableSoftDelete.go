package nuke

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
)

func disableSoftDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == azurenukeclient.AzureRecoveryVault && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureVault)
				var softDelete *armrecoveryservices.SoftDeleteState
				if item.Properties.SecuritySettings != nil && item.Properties.SecuritySettings.SoftDeleteSettings != nil {
					softDelete = item.Properties.SecuritySettings.SoftDeleteSettings.SoftDeleteState

				}

				if softDelete != nil && *softDelete != armrecoveryservices.SoftDeleteStateAlwaysON &&
					*softDelete != armrecoveryservices.SoftDeleteStateEnabled {
					continue
				}

				logger.Info(fmt.Sprintf("disableSoftDelete for vault :%s", *item.Name))
				err := state.azureClient.DisableSoftDelete(ctx, item.Vault)
				if err != nil {
					return composed.LogErrorAndReturn(err, fmt.Sprintf("Error Disabling SoftDelete on Azure Vault: %s", obj.GetId()), composed.StopWithRequeue, ctx)
				}
			}
		}

	}
	return nil, nil
}
