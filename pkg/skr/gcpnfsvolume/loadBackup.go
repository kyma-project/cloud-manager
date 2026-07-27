package gcpnfsvolume

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadBackup(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolume()

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if restore.Spec.SourceBackupUrl == "" {
		logger.Info("Skipping backup load as no BackupUrl is provided")
		return nil, ctx
	}

	// On a non-GCP scope createBackupClient is a no-op and leaves fileBackupClient nil;
	// stay consistent with that guard rather than dereferencing a nil Gcp scope/client.
	if state.Scope.Spec.Scope.Gcp == nil || state.fileBackupClient == nil {
		return nil, ctx
	}

	location := extractBackupLocation(restore.Spec.SourceBackupUrl)
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	name := extractBackupName(restore.Spec.SourceBackupUrl)

	backup, err := state.fileBackupClient.GetFilestoreBackup(ctx, &filestorepb.GetBackupRequest{
		Name: gcpnfsbackupclientv2.GetFileBackupPath(project, location, name),
	})

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

	return nil, ctx
}
