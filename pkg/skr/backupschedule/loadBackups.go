package backupschedule

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

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Load Backups")

	list := state.backupImpl.emptyBackupList()
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

	//convert list to a slice
	objects := state.backupImpl.toObjectSlice(list)

	//sort the objects
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].GetCreationTimestamp().Time.Before(objects[j].GetCreationTimestamp().Time)
	})

	//Store the objects in the State.
	state.Backups = objects

	return nil, nil
}
