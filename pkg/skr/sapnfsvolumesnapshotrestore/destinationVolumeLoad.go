package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func destinationVolumeLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	destRef := restore.Spec.Destination.ExistingVolume
	ns := destRef.Namespace
	if ns == "" {
		ns = restore.Namespace
	}

	nfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      destRef.Name,
		Namespace: ns,
	}, nfsVolume)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Destination SapNfsVolume not found", "name", destRef.Name, "namespace", ns)
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolume,
				Message: fmt.Sprintf("Destination SapNfsVolume %s/%s not found", ns, destRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Check if the volume is Ready
	volumeReady := meta.FindStatusCondition(nfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("destination SapNfsVolume %s/%s is not ready", ns, destRef.Name), "Destination volume not ready")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
				Message: fmt.Sprintf("Destination SapNfsVolume %s/%s is not ready", ns, destRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Resolve the KCP NfsInstance to get the Manila share ID
	if nfsVolume.Status.Id == "" {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("destination SapNfsVolume has no status ID"), "Destination volume has no status ID")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
				Message: "Destination SapNfsVolume has no status ID",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
	err = state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      nfsVolume.Status.Id,
	}, kcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance for destination volume", composed.StopWithRequeue, ctx)
	}

	shareId, _ := kcpNfsInstance.GetStateData("shareId")
	if shareId == "" {
		return composed.StopWithRequeue, ctx
	}

	state.DestinationVolume = nfsVolume
	state.shareId = shareId

	return nil, ctx
}
