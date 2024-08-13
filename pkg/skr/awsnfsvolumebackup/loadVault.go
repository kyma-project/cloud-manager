package awsnfsvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
	"time"
)

func loadVault(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, nil
	}

	if state.vault != nil {
		return nil, nil
	}

	vaultName := state.GetVaultName()
	vaults, err := state.awsClient.ListBackupVaults(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS Backup Vaults", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	//Match the vault by name. If found, continue...
	for _, vault := range vaults {
		if ptr.Deref(vault.BackupVaultName, "") == vaultName {
			state.vault = &vault
			return nil, nil
		}
	}

	// vault does not exist
	logger.Info("Creating AWS Backup Vault")
	_, err = state.awsClient.CreateBackupVault(ctx, vaultName, map[string]string{
		"cloud-resources.kyma-project.io/scope": state.Scope().Name,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS Backup Vault", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	// it should load the created one, hopefully won't end up in endless recursion, needs real test pass
	return loadVault(ctx, state)
}
