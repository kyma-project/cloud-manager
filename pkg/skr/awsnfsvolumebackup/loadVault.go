package awsnfsvolumebackup

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func loadVault(ctx context.Context, st composed.State, local bool) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	client := state.awsClient
	if !local {
		client = state.destAwsClient
	}

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, ctx
	}

	if state.vault != nil {
		return nil, ctx
	}

	// load backup vaults
	logger.WithValues("local", local).Info("Loading AWS Backup Vault")
	vaultName := state.GetVaultName()
	vaults, err := client.ListBackupVaults(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS Backup Vaults", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	//Match the vault by name. If found, continue...
	for _, vault := range vaults {
		if ptr.Deref(vault.BackupVaultName, "") == vaultName {
			if local {
				state.vault = &vault
			} else {
				state.destVault = &vault
			}
			return nil, ctx
		}
	}

	// vault does not exist
	logger.WithValues("local", local).Info("Creating AWS Backup Vault")
	_, err = state.awsClient.CreateBackupVault(ctx, vaultName, map[string]string{
		"cloud-resources.kyma-project.io/scope": state.Scope().Name,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS Backup Vault", err, ctx)
	}

	// it should load the created one, hopefully won't end up in endless recursion, needs real test pass
	return loadVault(ctx, state, local)
}

func loadLocalVault(ctx context.Context, st composed.State) (error, context.Context) {
	return loadVault(ctx, st, true)
}

func loadDestVault(ctx context.Context, st composed.State) (error, context.Context) {
	return loadVault(ctx, st, false)
}
