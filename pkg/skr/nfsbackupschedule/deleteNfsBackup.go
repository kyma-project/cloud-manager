package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Deleting old File Backups")

	temp := schedule.Status.Backups[:0]
	for _, ref := range schedule.Status.Backups {
		backup, err := getBackupObject(ctx, state, ref)
		if err != nil {
			logger.WithValues("Backup", ref.Name).Info("Error getting backup object")
			if !errors.IsNotFound(err) {
				//Adding it back to the list. Can try it next iteration
				temp = append(temp, ref)
			}
			continue
		}

		//Check if the backup object should be deleted
		toRetain := time.Duration(schedule.Spec.MaxRetentionDays) * 24 * time.Hour
		elapsed := time.Since(backup.GetCreationTimestamp().Time)
		if elapsed > toRetain {
			logger.WithValues("Backup", ref.Name).Info("Deleting backup object")
			err = state.Cluster().K8sClient().Delete(ctx, backup)
			if err == nil {
				continue
			}
		}
		temp = append(temp, ref)
	}

	//Update the status of the schedule with the list of available backups
	schedule.Status.Backups = temp
	schedule.Status.LastDeleteRun = &metav1.Time{Time: state.nextRunTime.UTC()}
	return composed.UpdateStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}

func getBackupObject(ctx context.Context, state *State, ref corev1.ObjectReference) (client.Object, error) {

	key := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}

	var obj client.Object

	switch state.Scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		obj = &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
	case cloudcontrolv1beta1.ProviderAws:
		obj = &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	default:
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}
	err := state.SkrCluster.K8sClient().Get(ctx, key, obj)
	return obj, err
}
