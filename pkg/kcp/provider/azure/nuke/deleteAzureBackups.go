package nuke

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteAzureBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "AzureRwxVolumeBackup" && rks.Provider == cloudcontrolv1beta1.ProviderAzure {
			for _, obj := range rks.Objects {

				item := obj.(azureProtectedItem)
				protected, okay := item.Properties.(*armrecoveryservicesbackup.AzureFileshareProtectedItem)
				if !okay {
					continue
				}

				if protected.ProtectionState != to.Ptr(armrecoveryservicesbackup.ProtectionStateProtected) {
					continue
				}

				err := state.azureClient.RemoveProtection(ctx, item.ProtectedItemResource)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Error requesting Azure File Backup deletion %s", obj.GetId()))
				}
			}
		}

	}
	return nil, nil
}
