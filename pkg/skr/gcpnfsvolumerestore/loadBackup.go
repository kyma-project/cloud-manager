package gcpnfsvolumerestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadBackup(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolumeRestore()

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if restore.Spec.Source.BackupUrl == "" {
		logger.Info("Skipping backup load as no BackupUrl is provided")
		return nil, nil
	}

	location := extractBackupLocation(restore.Spec.Source.BackupUrl)
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	name := extractBackupName(restore.Spec.Source.BackupUrl)

	backup, err := state.fileBackupClient.GetFileBackup(ctx, project, location, name)

	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.JobStateError,
				Message: "Failed to get backup details for permission check",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.fileBackup = backup

	return nil, nil
}
