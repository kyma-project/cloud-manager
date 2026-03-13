package v2

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	gcpnfsbackupclientv2util "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func populateBackupUrl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsGcpNfsVolumeRestore()
	logger.WithValues("nfsRestoreSource", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Loading GCPNfsVolumeBackup")

	if len(restore.Spec.Source.Backup.Name) > 0 {
		nfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		nfsVolumeBackupKey := restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace())
		err := state.SkrCluster.K8sClient().Get(ctx, nfsVolumeBackupKey, nfsVolumeBackup)
		if client.IgnoreNotFound(err) != nil {
			return composed.LogErrorAndReturn(err, "Error loading SKR GcpNfsVolumeRestore", composed.StopWithRequeue, ctx)
		}
		if err != nil {
			restore.Status.State = cloudresourcesv1beta1.JobStateError
			logger.Error(err, "Error getting GcpNfsVolumeBackup")
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolumeBackup,
					Message: "Error loading GcpNfsVolumeBackup",
				}).
				SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
				Run(ctx, state)
		}

		isBackupReady := meta.IsStatusConditionTrue(nfsVolumeBackup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

		if !isBackupReady {
			logger.WithValues("GcpNfsVolumeBackup", nfsVolumeBackup.Name).Info("GcpNfsVolumeBackup is not ready")
			restore.Status.State = cloudresourcesv1beta1.JobStateError
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
					Message: "GcpNfsVolumeBackup is not ready",
				}).
				SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
				SuccessLogMsg("Error loading GcpNfsVolumeBackup").
				Run(ctx, state)
		}

		gcpScope := state.Scope.Spec.Scope.Gcp
		project := gcpScope.Project
		srcLocation := nfsVolumeBackup.Status.Location
		backupName := fmt.Sprintf("cm-%.60s", nfsVolumeBackup.Status.Id)
		state.SrcBackupFullPath = gcpnfsbackupclientv2util.GetFileBackupPath(project, srcLocation, backupName)

		return nil, nil
	}

	project := state.Scope.Spec.Scope.Gcp.Project
	if restore.Spec.Source.BackupUrl != "" {
		// Convert Source.BackupUrl from {location_id}/{backup_id} to full GCP path
		fullPath, err := convertBackupUrlToFullPath(project, restore.Spec.Source.BackupUrl)
		if err != nil {
			restore.Status.State = cloudresourcesv1beta1.StateError
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: fmt.Sprintf("Invalid SourceBackupUrl format: %s", err.Error()),
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Error converting SourceBackupUrl format").
				Run(ctx, state)
		}
		state.SrcBackupFullPath = fullPath
		return nil, nil
	}

	err := fmt.Errorf("either Spec.Source.Backup.Name or Spec.Source.BackupUrl must be set")
	return composed.LogErrorAndReturn(err, "Invalid GcpNfsVolumeRestore specification", composed.StopWithRequeue, ctx)
}
