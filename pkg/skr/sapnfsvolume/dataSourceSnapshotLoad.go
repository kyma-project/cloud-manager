package sapnfsvolume

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func dataSourceSnapshotLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	vol := state.ObjAsSapNfsVolume()

	if vol.Spec.DataSource == nil || vol.Spec.DataSource.Snapshot == nil {
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(vol) {
		return nil, ctx
	}
	if state.KcpNfsInstance != nil {
		return nil, ctx
	}

	ref := vol.Spec.DataSource.Snapshot
	ns := ref.Namespace
	if ns == "" {
		ns = vol.Namespace
	}

	snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ns}, snapshot)
	if apierrors.IsNotFound(err) {
		vol.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(vol).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolumeBackup,
				Message: fmt.Sprintf("DataSource SapNfsVolumeSnapshot %s/%s not found", ns, ref.Name),
			}).
			ErrorLogMessage("Error patching SapNfsVolume status after missing dataSource snapshot").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading dataSource SapNfsVolumeSnapshot", composed.StopWithRequeue, ctx)
	}

	snapshotReady := meta.FindStatusCondition(snapshot.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if snapshotReady == nil || snapshotReady.Status != metav1.ConditionTrue {
		vol.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(vol).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
				Message: fmt.Sprintf("DataSource SapNfsVolumeSnapshot %s/%s is not ready", ns, ref.Name),
			}).
			ErrorLogMessage("Error patching SapNfsVolume status after not-ready dataSource snapshot").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if snapshot.Status.OpenstackId == "" {
		vol.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(vol).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
				Message: fmt.Sprintf("DataSource SapNfsVolumeSnapshot %s/%s has no openstackId", ns, ref.Name),
			}).
			ErrorLogMessage("Error patching SapNfsVolume status after dataSource snapshot missing openstackId").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.DataSourceSnapshot = snapshot

	return nil, ctx
}
