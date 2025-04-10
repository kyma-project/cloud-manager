package azurerwxvolumerestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadAzureRwxVolumeBackup(ctx context.Context, st composed.State) (error, context.Context) {
	// implement similar to loadAzureRwxVolume
	// loadAzureRwxVolumeBackup loads the AzureRwxVolumeBackup object from the restore.Spec.Source.Backup value.
	// If the object is not found, the function returns an error.
	// If the object is found but is not ready, the function returns an error.
	// If the object is found and is ready, the function stores the object in the state.
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsAzureRwxVolumeRestore()
	logger = logger.WithValues("AzureRwxVolumeRestoreSource", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination", restore.Spec.Destination.Pvc.ToNamespacedName(state.Obj().GetNamespace()))
	composed.LoggerIntoCtx(ctx, logger)
	logger.Info("Loading AzureRwxVolumeBackup")
	//Load the rwxVolumeBackup object
	rwxVolumeBackup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}
	rwxVolumeBackupKey := restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace())
	err := state.Cluster().K8sClient().Get(ctx, rwxVolumeBackupKey, rwxVolumeBackup)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR AzureRwxVolumeRestore", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		logger.Error(err, "Error getting AzureRwxVolumeBackup")
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingRwxVolumeBackup,
				Message: "Error loading AzureRwxVolumeBackup",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	//Check if the azureRwxVolumeBackup has a ready condition
	backupReady := meta.FindStatusCondition(rwxVolumeBackup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	//If the rwxVolume is not ready, return an error
	if backupReady == nil || backupReady.Status != metav1.ConditionTrue {
		logger.WithValues("AzureRwxVolumeBackup", rwxVolumeBackup.Name).Info("AzureRwxVolumeBackup is not ready")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonRwxVolumeBackupNotReady,
				Message: "AzureRwxVolumeBackup is not ready",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	//Store the azureRwxVolumeBackup in state
	state.azureRwxVolumeBackup = rwxVolumeBackup

	return nil, nil
}
