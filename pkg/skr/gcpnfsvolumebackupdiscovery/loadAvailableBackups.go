package gcpnfsvolumebackupdiscovery

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
)

func loadAvailableBackups(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	backups, err := state.fileBackupClient.ListFilesBackups(
		ctx, state.Scope.Spec.Scope.Gcp.Project,
		gcpclient.GetSharedBackupsFilter(state.Scope.Spec.ShootName, state.Scope.Spec.ShootName), // todo: change second arg to SubaccountId
	)

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("No shared backups found for this cluster")
			return nil, ctx
		}

		backupDiscovery.Status.State = cloudresourcesv1beta1.StateError
		errMsg := "Failed to load shared Nfs Volume Backups from GCP"
		return composed.UpdateStatus(backupDiscovery).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage(errMsg).
			SuccessLogMsg("Updated and forgot SKR GcpNfsVolumeBackupDiscovery status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.backups = backups

	return nil, ctx
}
