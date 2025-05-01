package azurerwxvolumebackup

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultPolicyName = "cm-noop-policy"

func createBackupPolicy(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()

	// short circuit if protectedResourceName already exist
	if state.protectedResourceName != "" {
		return nil, ctx
	}

	vaultName := state.vaultName
	resourceGroupName := state.resourceGroupName

	err := state.client.CreateBackupPolicy(ctx, vaultName, resourceGroupName, DefaultPolicyName)
	if err != nil {
		logger.Error(err, "failed to create backup policy")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: fmt.Sprintf("Could not create BackupPolicy for backup: %s", err),
			}).
			FailedError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Success; continue on
	return nil, ctx
}
