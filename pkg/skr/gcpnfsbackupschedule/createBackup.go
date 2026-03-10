package gcpnfsbackupschedule

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	if state.Source == nil {
		logger.Error(fmt.Errorf("source not loaded"), "Source is nil")
		return composed.StopWithRequeue, nil
	}

	if state.createRunDone {
		logger.Info(fmt.Sprintf("Creation already completed for %s ", state.nextRunTime))
		return nil, nil
	}

	if state.Scheduler.GetRemainingTime(state.nextRunTime) > 0 {
		return nil, nil
	}

	logger.Info("Creating File Backup")

	// Get Backup details (name and index)
	prefix := schedule.GetPrefix()
	if prefix == "" {
		prefix = schedule.GetName()
	}
	index := schedule.GetBackupIndex() + 1
	name := fmt.Sprintf("%s-%d-%s", prefix, index, state.nextRunTime.UTC().Format("20060102-150405"))

	gcpSchedule := state.ObjAsGcpNfsBackupSchedule()

	backup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: schedule.GetNamespace(),
			Labels: map[string]string{
				cloudresourcesv1beta1.LabelScheduleName:      schedule.GetName(),
				cloudresourcesv1beta1.LabelScheduleNamespace: schedule.GetNamespace(),
			},
		},
		Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
			Location: gcpSchedule.Spec.Location,
			Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
				Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
					Name:      state.Source.Name,
					Namespace: state.Source.Namespace,
				},
			},
			AccessibleFrom: gcpSchedule.Spec.AccessibleFrom,
		},
	}

	// Check if backup already exists before creating
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      backup.Name,
		Namespace: backup.Namespace,
	}, &cloudresourcesv1beta1.GcpNfsVolumeBackup{})
	if err != nil && apierrors.IsNotFound(err) {
		err = state.Cluster().K8sClient().Create(ctx, backup)
	}

	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg(fmt.Sprintf("Error creating GcpNfsVolumeBackup: %s", err)).
			Run(ctx, state)
	}

	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetBackupIndex(index)
	schedule.SetBackupCount(len(state.Backups) + 1)
	schedule.SetLastCreateRun(&metav1.Time{Time: state.nextRunTime.UTC()})
	schedule.SetLastCreatedBackup(corev1.ObjectReference{
		Kind:      "GcpNfsVolumeBackup",
		Namespace: backup.Namespace,
		Name:      backup.Name,
	})
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
