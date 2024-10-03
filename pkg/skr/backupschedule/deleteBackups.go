package backupschedule

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	//If marked for deletion, return
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//Check next run time. If it is not time to run, return
	if GetRemainingTimeFromNow(&state.nextRunTime) > 0 {
		return nil, nil
	}

	//If the deletion for the nextRunTime is already done, return
	if schedule.GetLastDeleteRun() != nil && !schedule.GetLastDeleteRun().IsZero() &&
		GetRemainingTime(&state.nextRunTime, &schedule.GetLastDeleteRun().Time) == 0 {
		logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Deletion already completed for %s ", state.nextRunTime))
		return nil, nil
	}

	//If the number of backups is zero, OR
	//If maxRetentionDays is not positive, requeue to update next run time
	if len(state.Backups) == 0 || schedule.GetMaxRetentionDays() <= 0 {
		schedule.SetLastDeleteRun(&metav1.Time{Time: state.nextRunTime.UTC()})
		schedule.SetNextDeleteTimes(nil)
		schedule.SetLastDeletedBackups(nil)
		schedule.SetBackupCount(len(state.Backups))
		return composed.PatchStatus(schedule).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info("Deleting old File Backups")

	nextDeleteTimes := map[string]string{}
	var lastDeleted []corev1.ObjectReference
	for _, backup := range state.Backups {

		//Check if the backup object should be deleted
		toRetain := time.Duration(schedule.GetMaxRetentionDays()) * 24 * time.Hour
		elapsed := time.Since(backup.GetCreationTimestamp().Time)
		if elapsed > toRetain {
			logger.WithValues("Backup", backup.GetName()).Info("Deleting backup object")
			err := state.Cluster().K8sClient().Delete(ctx, backup)
			if err != nil {
				logger.Error(err, "Error deleting the backup object.")
				continue
			}
			lastDeleted = append(lastDeleted, corev1.ObjectReference{
				Name:      backup.GetName(),
				Namespace: backup.GetNamespace(),
			})
		}
		if len(nextDeleteTimes) < MaxSchedules {
			backupName := fmt.Sprintf("%s/%s", backup.GetNamespace(), backup.GetName())
			deleteTime := backup.GetCreationTimestamp().AddDate(0, 0, schedule.GetMaxRetentionDays())
			nextDeleteTimes[backupName] = deleteTime.UTC().Format(time.RFC3339)
		}
	}

	//Update the status of the schedule with the list of available backups
	schedule.SetLastDeleteRun(&metav1.Time{Time: state.nextRunTime.UTC()})
	schedule.SetLastDeletedBackups(lastDeleted)
	schedule.SetNextDeleteTimes(nextDeleteTimes)
	schedule.SetBackupCount(len(state.Backups) - len(lastDeleted))
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
