package sapnfsvolumesnapshot

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

func sourceVolumeLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If marked for deletion, skip source validation
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	// If snapshot already loaded (Manila snapshot exists), skip
	if state.snapshot != nil {
		return nil, ctx
	}

	// Load the SapNfsVolume
	sourceRef := snapshot.Spec.SourceVolume
	ns := sourceRef.Namespace
	if ns == "" {
		ns = snapshot.Namespace
	}

	nfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      sourceRef.Name,
		Namespace: ns,
	}, nfsVolume)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Source SapNfsVolume not found", "name", sourceRef.Name, "namespace", ns)
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolume,
				Message: fmt.Sprintf("Source SapNfsVolume %s/%s not found", ns, sourceRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Check if the volume is Ready
	volumeReady := meta.FindStatusCondition(nfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("source SapNfsVolume %s/%s is not ready", ns, sourceRef.Name), "Source volume not ready")
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
				Message: fmt.Sprintf("Source SapNfsVolume %s/%s is not ready", ns, sourceRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Resolve the KCP NfsInstance to get the Manila share ID
	if nfsVolume.Status.Id == "" {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("source SapNfsVolume has no status ID"), "Source volume has no status ID")
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
				Message: "Source SapNfsVolume has no status ID",
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
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance for source volume", composed.StopWithRequeue, ctx)
	}

	shareId, _ := kcpNfsInstance.GetStateData("shareId")
	if shareId == "" {
		return composed.StopWithRequeue, ctx
	}

	// Store the share ID in status if not already set
	if snapshot.Status.ShareId != shareId {
		snapshot.Status.ShareId = shareId
		err = composed.PatchObjStatus(ctx, snapshot, state.Cluster().K8sClient())
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error patching snapshot status with shareId", composed.StopWithRequeue, ctx)
		}
	}

	state.SapNfsVolume = nfsVolume

	return nil, ctx
}
