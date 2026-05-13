package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func markFailed(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// If deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// Only transition from Error to Failed
	if snapshot.Status.State != cloudresourcesv1beta1.StateError {
		return nil, ctx
	}

	// If this snapshot doesn't belong to a schedule, continue
	scheduleName, exists := snapshot.GetLabels()[cloudresourcesv1beta1.LabelScheduleName]
	if !exists {
		return nil, ctx
	}

	// List subsequent snapshots for this schedule
	list := &cloudresourcesv1beta1.SapNfsVolumeSnapshotList{}
	err := state.SkrCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			cloudresourcesv1beta1.LabelScheduleName:      scheduleName,
			cloudresourcesv1beta1.LabelScheduleNamespace: snapshot.GetNamespace(),
		},
		client.InNamespace(snapshot.GetNamespace()),
	)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Error listing subsequent snapshots")
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonBackupListFailed,
				Message: fmt.Sprintf("Error listing subsequent snapshot(s): %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// If there are subsequent snapshots, mark this one as failed
	for _, item := range list.Items {
		if item.CreationTimestamp.After(snapshot.CreationTimestamp.Time) {
			errCondition := meta.FindStatusCondition(snapshot.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			message := "Snapshot moved to Failed state as more recent snapshot(s) is available."
			if errCondition != nil {
				message = errCondition.Message + "\n" + message
			}

			logger := composed.LoggerFromCtx(ctx)
			logger.Error(fmt.Errorf("snapshot superseded by newer snapshot"), "Marking snapshot as failed")
			snapshot.Status.State = cloudresourcesv1beta1.StateFailed
			return composed.PatchStatus(snapshot).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ReasonBackupFailed,
					Message: message,
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
	}

	return nil, ctx
}
