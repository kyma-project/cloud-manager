package backupschedule

import (
	"context"
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
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

	//If the creation for the nextRunTime is already done, return
	if schedule.GetLastCreateRun() != nil && !schedule.GetLastCreateRun().IsZero() &&
		state.nextRunTime.Unix() == schedule.GetLastCreateRun().Time.Unix() {
		logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Creation already completed for %s ", state.nextRunTime))
		return nil, nil
	}

	logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info("Creating File Backup")

	//Get Backup details (name and index).
	prefix := schedule.GetPrefix()
	if prefix == "" {
		prefix = schedule.GetName()
	}
	index := schedule.GetBackupIndex() + 1
	name := fmt.Sprintf("%s-%d-%s", prefix, index, state.nextRunTime.UTC().Format("20060102-150405"))

	//Construct a common part of Backup Object
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: schedule.GetNamespace(),
		Labels: map[string]string{
			cloudresourcesv1beta1.LabelScheduleName:      schedule.GetName(),
			cloudresourcesv1beta1.LabelScheduleNamespace: schedule.GetNamespace(),
		},
	}

	//Create Backup object
	backup, err := state.backupImpl.getBackupObject(state, &objectMeta)
	if err == nil {
		err = state.Cluster().K8sClient().Create(ctx, backup)
	}

	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg(fmt.Sprintf("Error updating File backup object in GCP :%s", err)).
			Run(ctx, state)
	}

	//Update the status of the schedule with the latest backup details
	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetBackupIndex(index)
	schedule.SetBackupCount(len(state.Backups) + 1)
	schedule.SetLastCreateRun(&metav1.Time{Time: state.nextRunTime.UTC()})
	schedule.SetLastCreatedBackup(corev1.ObjectReference{
		Kind:      backup.GetObjectKind().GroupVersionKind().Kind,
		Namespace: backup.GetNamespace(),
		Name:      backup.GetName(),
	})
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
