package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsNfsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, return
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//Check next run time. If it is not time to run, return
	if state.nextRunTime.IsZero() || now.Before(state.nextRunTime) {
		return nil, nil
	}

	//If the deletion for the nextRunTime is already done, return
	if schedule.Status.LastDeleteRun != nil && !schedule.Status.LastDeleteRun.IsZero() &&
		state.nextRunTime.Unix() == schedule.Status.LastDeleteRun.Time.Unix() {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info(fmt.Sprintf("Deletion already completed for %s ", state.nextRunTime))
		return nil, nil
	}

	//If maxRetentionDays is not positive, requeue to update next run time
	if schedule.Spec.MaxRetentionDays <= 0 {
		schedule.Status.LastDeleteRun = &metav1.Time{Time: state.nextRunTime.UTC()}
		schedule.Status.NextDeleteTimes = nil
		schedule.Status.LastDeletedBackups = nil
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Deleting old File Backups")

	list := &cloudresourcesv1beta1.GcpNfsVolumeBackupList{}
	err := state.SkrCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			cloudresourcesv1beta1.LabelScheduleName:      schedule.Name,
			cloudresourcesv1beta1.LabelScheduleNamespace: schedule.Namespace,
		},
		client.InNamespace(schedule.Namespace),
	)

	if err != nil {
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonBackupListFailed,
				Message: "Error listing backup(s)",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//sort the list based on creationTime.
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].CreationTimestamp.Before(&list.Items[j].CreationTimestamp)
	})

	nextDeleteTimes := map[string]string{}
	var lastDeleted []corev1.ObjectReference
	for _, backup := range list.Items {

		//Check if the backup object should be deleted
		toRetain := time.Duration(schedule.Spec.MaxRetentionDays) * 24 * time.Hour
		elapsed := time.Since(backup.GetCreationTimestamp().Time)
		if elapsed > toRetain {
			logger.WithValues("Backup", backup.Name).Info("Deleting backup object")
			err = state.Cluster().K8sClient().Delete(ctx, &backup)
			if err != nil {
				logger.Error(err, "Error deleting the backup object.")
				continue
			}
			lastDeleted = append(lastDeleted, corev1.ObjectReference{
				Name:      backup.Name,
				Namespace: backup.Namespace,
			})
		}
		if len(nextDeleteTimes) < MaxSchedules {
			backupName := fmt.Sprintf("%s/%s", backup.Namespace, backup.Name)
			deleteTime := backup.CreationTimestamp.AddDate(0, 0, schedule.Spec.MaxRetentionDays)
			nextDeleteTimes[backupName] = deleteTime.UTC().Format(time.RFC3339)
		}
	}

	//Update the status of the schedule with the list of available backups
	//schedule.Status.Backups = temp
	schedule.Status.LastDeleteRun = &metav1.Time{Time: state.nextRunTime.UTC()}
	schedule.Status.LastDeletedBackups = lastDeleted
	schedule.Status.NextDeleteTimes = nextDeleteTimes
	return composed.UpdateStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
