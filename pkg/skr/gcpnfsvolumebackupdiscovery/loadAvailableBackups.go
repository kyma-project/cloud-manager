package gcpnfsvolumebackupdiscovery

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
)

func loadAvailableBackups(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	project := state.Scope.Spec.Scope.Gcp.Project
	filter := gcpclient.GetSharedBackupsFilter(state.Scope.Spec.ShootName, state.Scope.Spec.ShootName) // todo: change second arg to SubaccountId
	iter := state.fileBackupClient.ListFilestoreBackups(ctx, &filestorepb.ListBackupsRequest{
		Parent: v2client.GetFilestoreParentPath(project, "-"),
		Filter: filter,
	})
	var backups []*filestorepb.Backup
	var err error
	for b, iterErr := range iter.All() {
		if iterErr != nil {
			err = iterErr
			break
		}
		backups = append(backups, b)
	}

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
