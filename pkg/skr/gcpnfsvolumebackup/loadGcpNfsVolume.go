package gcpnfsvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadGcpNfsVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If marked for deletion, return
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If the GCP backup already exists, return
	if state.fileBackup != nil {
		return nil, nil
	}

	backup := state.ObjAsGcpNfsVolumeBackup()
	logger.WithValues("Nfs Backup :", backup.Name).Info("Loading GCPNfsVolume")

	//Load the nfsVolume object
	nfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
	nfsVolumeKey := types.NamespacedName{
		Name:      backup.Spec.Source.Volume.Name,
		Namespace: backup.Spec.Source.Volume.Namespace,
	}
	err := state.SkrCluster.K8sClient().Get(ctx, nfsVolumeKey, nfsVolume)
	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.UpdateStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: "Error loading GcpNfsVolume",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting GcpNfsVolume").
			Run(ctx, state)
	}

	//Check if the nfsVolume has a ready condition
	volumeReady := meta.FindStatusCondition(nfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	//If the nfsVolume is not ready, return an error
	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger.WithValues("GcpNfsVolume", nfsVolume.Name).Info("GcpNfsVolume is ready")
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.UpdateStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: "Error loading GcpNfsVolume",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting GcpNfsVolume").
			Run(ctx, state)
	}

	//Store the gcpNfsVolume in state
	state.GcpNfsVolume = nfsVolume

	return nil, nil
}
