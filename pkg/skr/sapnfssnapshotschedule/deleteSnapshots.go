package sapnfssnapshotschedule

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteSnapshots(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	// If the deletion for the nextRunTime is already done, return
	if state.deleteRunDone {
		logger.Info(fmt.Sprintf("Deletion already completed for %s ", state.nextRunTime))
		return nil, ctx
	}

	// Check next run time. If it is not time to run, return
	if state.Scheduler.GetRemainingTime(state.nextRunTime) > 0 {
		return nil, ctx
	}

	// If the number of snapshots is zero, OR maxRetentionDays is not positive, requeue to update next run time
	if len(state.Snapshots) == 0 || schedule.GetMaxRetentionDays() <= 0 {
		schedule.SetLastDeleteRun(&metav1.Time{Time: state.nextRunTime.UTC()})
		schedule.SetNextDeleteTimes(nil)
		schedule.SetLastDeletedBackups(nil)
		schedule.SetBackupCount(len(state.Snapshots))
		return composed.PatchStatus(schedule).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.Info("Deleting old NFS Volume Snapshots")

	nextDeleteTimes := map[string]string{}
	var lastDeleted []corev1.ObjectReference
	readyCount, failedCount := 0, 0

	for _, snapshot := range state.Snapshots {
		// Check if the snapshot object should be deleted
		toRetain := time.Duration(schedule.GetMaxRetentionDays()) * 24 * time.Hour
		elapsed := time.Since(snapshot.GetCreationTimestamp().Time)
		if elapsed > toRetain ||
			(snapshot.Status.State == cloudresourcesv1beta1.StateReady && readyCount >= schedule.GetMaxReadyBackups()) ||
			(snapshot.Status.State == cloudresourcesv1beta1.StateFailed && failedCount >= schedule.GetMaxFailedBackups()) {
			logger.WithValues("Snapshot", snapshot.GetName()).Info("Deleting snapshot object")
			err := state.Cluster().K8sClient().Delete(ctx, snapshot)
			if err != nil {
				logger.Error(err, "Error deleting the snapshot object.")
				continue
			}
			lastDeleted = append(lastDeleted, corev1.ObjectReference{
				Name:      snapshot.GetName(),
				Namespace: snapshot.GetNamespace(),
			})
		} else {
			switch snapshot.Status.State {
			case cloudresourcesv1beta1.StateReady:
				readyCount++
			case cloudresourcesv1beta1.StateFailed:
				failedCount++
			}
		}
		if uint(len(nextDeleteTimes)) < backupschedule.MaxSchedules {
			backupName := fmt.Sprintf("%s/%s", snapshot.GetNamespace(), snapshot.GetName())
			deleteTime := snapshot.GetCreationTimestamp().AddDate(0, 0, schedule.GetMaxRetentionDays())
			nextDeleteTimes[backupName] = deleteTime.UTC().Format(time.RFC3339)
		}
	}

	// Update the status of the schedule
	schedule.SetLastDeleteRun(&metav1.Time{Time: state.nextRunTime.UTC()})
	schedule.SetLastDeletedBackups(lastDeleted)
	schedule.SetNextDeleteTimes(nextDeleteTimes)
	schedule.SetBackupCount(len(state.Snapshots) - len(lastDeleted))
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
