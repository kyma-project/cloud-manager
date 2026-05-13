package sapnfssnapshotschedule

import (
	"context"
	"sort"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSnapshots(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	list := &cloudresourcesv1beta1.SapNfsVolumeSnapshotList{}
	err := state.SkrCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			cloudresourcesv1beta1.LabelScheduleName:      schedule.GetName(),
			cloudresourcesv1beta1.LabelScheduleNamespace: schedule.GetNamespace(),
		},
		client.InNamespace(schedule.GetNamespace()),
	)

	if err != nil {
		logger.Error(err, "Error listing snapshots.")
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonBackupListFailed,
				Message: "Error listing snapshot(s)",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Convert to typed slice
	snapshots := make([]*cloudresourcesv1beta1.SapNfsVolumeSnapshot, 0, len(list.Items))
	for i := range list.Items {
		snapshots = append(snapshots, &list.Items[i])
	}

	// Sort in reverse chronological order
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].GetCreationTimestamp().After(snapshots[j].GetCreationTimestamp().Time)
	})

	state.Snapshots = snapshots

	return nil, ctx
}
