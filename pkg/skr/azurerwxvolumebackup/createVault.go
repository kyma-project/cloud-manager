package azurerwxvolumebackup

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"slices"
)

func vaultExists(vaults []*armrecoveryservices.Vault, location string) bool {

	return slices.ContainsFunc(vaults, func(vault *armrecoveryservices.Vault) bool {

		if vault.Location == nil || vault.Tags == nil {
			return false
		}

		_, tagExists := vault.Tags["cloud-manager"]

		return *vault.Location == location && tagExists

	})

}

func createVault(ctx context.Context, st composed.State) (error, context.Context) {

	// Read state
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()

	logger.WithValues("RwxBackup", backup.Name).Info("Creating Recovery Services Vault")

	location := backup.Spec.Location
	vaultName := fmt.Sprintf("cm-vault-%s", location)

	vaults, err := state.client.ListVaults(ctx)
	if err != nil {
		return composed.StopWithRequeue, ctx
	}

	state.vaultName = vaultName

	// If exists, exit early and go next action
	if vaultExists(vaults, location) {
		return nil, ctx
	}

	// TODO: resp gives the jobId. Use to check status
	resourceGroupName := state.resourceGroupName
	_, err = state.client.CreateVault(ctx, resourceGroupName, vaultName, location)
	if err != nil {
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx

}
