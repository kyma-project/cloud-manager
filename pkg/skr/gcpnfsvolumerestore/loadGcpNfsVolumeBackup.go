package gcpnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadGcpNfsVolumeBackup(ctx context.Context, st composed.State) (error, context.Context) {
	// implement similar to loadGcpNfsVolume
	// loadGcpNfsVolumeBackup loads the GcpNfsVolumeBackup object from the restore.Spec.Source.Backup value.
	// If the object is not found, the function returns an error.
	// If the object is found but is not ready, the function returns an error.
	// If the object is found and is ready, the function stores the object in the state.
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsGcpNfsVolumeRestore()
	logger.WithValues("Nfs Restore source:", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination:", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Loading GCPNfsVolumeBackup")

	//Load the nfsVolumeBackup object
	nfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
	nfsVolumeBackupKey := restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace())
	err := state.SkrCluster.K8sClient().Get(ctx, nfsVolumeBackupKey, nfsVolumeBackup)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR GcpNfsVolumeRestore", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		return composed.UpdateStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolumeBackup,
				Message: "Error loading GcpNfsVolumeBackup",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting GcpNfsVolumeBackup").
			Run(ctx, state)
	}

	//Check if the gcpNfsVolumeBackup has a ready condition
	backupReady := meta.FindStatusCondition(nfsVolumeBackup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	//If the nfsVolume is not ready, return an error
	if backupReady == nil || backupReady.Status != metav1.ConditionTrue {
		logger.WithValues("GcpNfsVolumeBackup", nfsVolumeBackup.Name).Info("GcpNfsVolumeBackup is not ready")
		return composed.UpdateStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
				Message: "GcpNfsVolumeBackup is not ready",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error loading GcpNfsVolumeBackup").
			Run(ctx, state)
	}

	//Store the gcpNfsVolume in state
	state.GcpNfsVolumeBackup = nfsVolumeBackup

	return nil, nil
}
