package gcpnfsvolumerestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadGcpNfsVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If deleting, continue with next steps.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	restore := state.ObjAsGcpNfsVolumeRestore()
	logger.WithValues("nfsRestoreSource", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Loading GCPNfsVolume")

	//Load the nfsVolume object
	nfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
	nfsVolumeKey := restore.Spec.Destination.Volume.ToNamespacedName(restore.Namespace)
	err := state.SkrCluster.K8sClient().Get(ctx, nfsVolumeKey, nfsVolume)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR GcpNfsVolume", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		logger.Error(err, "Error getting GcpNfsVolume")
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolume,
				Message: "Error loading GcpNfsVolume",
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}
	//If deleting, we still need the gcpNfsVolume object for finding the restore operation if it exists.
	if !composed.MarkedForDeletionPredicate(ctx, st) {

		//Check if the nfsVolume has a ready condition
		volumeReady := meta.FindStatusCondition(nfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

		//If the nfsVolume is not ready, return an error
		if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
			logger.WithValues("GcpNfsVolume", nfsVolume.Name).Info("GcpNfsVolume is not ready")
			restore.Status.State = cloudresourcesv1beta1.JobStateError
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
					Message: "Error loading GcpNfsVolume",
				}).
				SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
				SuccessLogMsg("Error getting GcpNfsVolume").
				Run(ctx, state)
		}
	}
	//Store the gcpNfsVolume in state
	state.GcpNfsVolume = nfsVolume

	return nil, nil
}
