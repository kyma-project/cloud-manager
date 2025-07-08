package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func loadVault(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//Get the vaultName and load it.
	vaultName := state.GetVaultName()
	vaults, err := state.awsClient.ListBackupVaults(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS Backup Vaults", err, ctx)
	}

	//Match the vault by name. If found, continue...
	for _, vault := range vaults {
		if ptr.Deref(vault.BackupVaultName, "") == vaultName {
			state.vault = &vault
		}
	}
	return nil, nil
}
