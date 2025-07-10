package azurerwxpv

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadAzureRecoveryVaults(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Load AzureRecoveryVaults")
	vaults, err := state.client.ListRwxVolumeBackupVaults(ctx, state.Scope().Spec.ShootName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error loading azure recovery vaults", err, ctx)
	}

	//Store it in state
	state.recoveryVaults = vaults
	logger.Info(fmt.Sprintf("Loaded Vaults :%d", len(state.recoveryVaults)))
	return nil, ctx
}
