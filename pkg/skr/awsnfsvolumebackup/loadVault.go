package awsnfsvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func loadVault(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vault != nil {
		return nil, nil
	}

	vault, err := state.awsClient.DescribeBackupVault(ctx, state.Scope().Name)
	if err == nil {
		state.vault = vault
		return nil, nil
	}
	if !state.awsClient.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error loading AWS Backup Vault", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	// vault does not exist

	logger.Info("Creating AWS Backup Vault")
	_, err = state.awsClient.CreateBackupVault(ctx, state.Scope().Name, map[string]string{
		"cloud-resources.kyma-project.io/scope": state.Scope().Name,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS Backup Vault", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	// it should load the created one, hopefully won't end up in endless recursion, needs real test pass
	return loadVault(ctx, state)
}
