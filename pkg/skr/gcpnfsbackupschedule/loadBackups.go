package gcpnfsbackupschedule

import (
	"context"
	"sort"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	list := &cloudresourcesv1beta1.GcpNfsVolumeBackupList{}
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
		logger.Error(err, "Error listing backups.")
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonBackupListFailed,
				Message: "Error listing backup(s)",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Convert to typed slice
	backups := make([]*cloudresourcesv1beta1.GcpNfsVolumeBackup, 0, len(list.Items))
	for i := range list.Items {
		backups = append(backups, &list.Items[i])
	}

	// Sort in reverse chronological order
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].GetCreationTimestamp().After(backups[j].GetCreationTimestamp().Time)
	})

	state.Backups = backups

	return nil, nil
}
