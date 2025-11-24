package gcpnfsvolume

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkRestorePermissions(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolume()

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if restore.Spec.SourceBackupUrl == "" {
		logger.Info("Skipping backup load as no BackupUrl is provided")
		return nil, nil
	}

	if !state.IsAllowedToRestoreBackup() {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.JobStateError,
				Message: "Not allowed to restore from the specified backup",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, nil
}
