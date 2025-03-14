package azurerwxvolumebackup

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func createVault(ctx context.Context, st composed.State) (error, context.Context) {

	// Read state
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()

	logger.WithValues("RwxBackup", backup.Name).Info("Creating Recovery Services Vault")

	resourceGroupName := "resourceGroupName"
	location := backup.Spec.Location
	vaultName := fmt.Sprintf("cm-vault-%s", location)

	// TODO: resp gives the jobId. Use to check status
	_, err := state.client.CreateVault(ctx, resourceGroupName, vaultName, location)
	if err != nil {
		return composed.StopWithRequeue, ctx
	}

	return composed.UpdateStatus(backup).
		ErrorLogMessage("").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)

}
