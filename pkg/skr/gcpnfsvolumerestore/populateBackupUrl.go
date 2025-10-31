package gcpnfsvolumerestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func populateBackupUrl(ctx context.Context, st composed.State) (error, context.Context) {
	//If deleting, continue with next steps.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	// implement similar to loadGcpNfsVolume
	// populateBackupUrl loads the GcpNfsVolumeBackup object from the restore.Spec.Source.Backup value.
	// If the object is not found, the function returns an error.
	// If the object is found but is not ready, the function returns an error.
	// If the object is found and is ready, the function stores the object in the state.
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsGcpNfsVolumeRestore()
	logger.WithValues("nfsRestoreSource", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Loading GCPNfsVolumeBackup")

	if len(restore.Spec.Source.Backup.Name) > 0 {
		//Load the nfsVolumeBackup object
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
				SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
				Run(ctx, state)
		}

		//Check if the gcpNfsVolumeBackup has a ready condition
		backupReady := meta.FindStatusCondition(nfsVolumeBackup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

		//If the nfsVolume is not ready, return an error
		if backupReady == nil || backupReady.Status != metav1.ConditionTrue {
			logger.WithValues("GcpNfsVolumeBackup", nfsVolumeBackup.Name).Info("GcpNfsVolumeBackup is not ready")
			restore.Status.State = cloudresourcesv1beta1.JobStateError
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
					Message: "GcpNfsVolumeBackup is not ready",
				}).
				SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
				SuccessLogMsg("Error loading GcpNfsVolumeBackup").
				Run(ctx, state)
		}

		gcpScope := state.Scope.Spec.Scope.Gcp
		project := gcpScope.Project
		srcLocation := nfsVolumeBackup.Status.Location
		backupName := fmt.Sprintf("cm-%.60s", nfsVolumeBackup.Status.Id)
		state.SrcBackupFullPath = gcpclient.GetFileBackupPath(project, srcLocation, backupName)

		return nil, nil
	}

	if restore.Spec.Source.BackupUrl != "" {
		// Use the provided backup URL directly
		state.SrcBackupFullPath = restore.Spec.Source.BackupUrl
		return nil, nil
	}

	err := fmt.Errorf("either Spec.Source.Backup.Name or Spec.Source.BackupUrl must be set")
	logger.Error(err, "Invalid GcpNfsVolumeRestore specification")
	return composed.LogErrorAndReturn(err, "Invalid GcpNfsVolumeRestore specification", composed.StopWithRequeue, ctx)
}
