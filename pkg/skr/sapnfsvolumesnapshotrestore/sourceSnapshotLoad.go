package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func sourceSnapshotLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	sourceRef := restore.Spec.SourceSnapshot
	ns := sourceRef.Namespace
	if ns == "" {
		ns = restore.Namespace
	}

	snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      sourceRef.Name,
		Namespace: ns,
	}, snapshot)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Source SapNfsVolumeSnapshot not found", "name", sourceRef.Name, "namespace", ns)
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolumeBackup,
				Message: fmt.Sprintf("Source SapNfsVolumeSnapshot %s/%s not found", ns, sourceRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Validate snapshot is Ready
	snapshotReady := meta.FindStatusCondition(snapshot.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if snapshotReady == nil || snapshotReady.Status != metav1.ConditionTrue {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("source SapNfsVolumeSnapshot %s/%s is not ready", ns, sourceRef.Name), "Source snapshot not ready")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
				Message: fmt.Sprintf("Source SapNfsVolumeSnapshot %s/%s is not ready", ns, sourceRef.Name),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if snapshot.Status.OpenstackId == "" {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("source SapNfsVolumeSnapshot has no openstackId"), "Source snapshot has no openstackId")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
				Message: "Source SapNfsVolumeSnapshot has no openstackId",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.SourceSnapshot = snapshot

	return nil, ctx
}
