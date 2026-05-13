package sapnfssnapshotschedule

import (
	"context"
	"fmt"
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createSnapshot(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	if state.Source == nil {
		logger.Error(fmt.Errorf("source not loaded"), "Source is nil")
		return composed.StopWithRequeue, ctx
	}

	if state.createRunDone {
		logger.Info(fmt.Sprintf("Creation already completed for %s ", state.nextRunTime))
		return nil, ctx
	}

	if state.Scheduler.GetRemainingTime(state.nextRunTime) > 0 {
		return nil, ctx
	}

	logger.Info("Creating NFS Volume Snapshot")

	// Get snapshot details (name and index)
	prefix := schedule.GetPrefix()
	if prefix == "" {
		prefix = schedule.GetName()
	}
	index := schedule.GetBackupIndex() + 1
	name := fmt.Sprintf("%s-%d-%s", prefix, index, state.nextRunTime.UTC().Format("20060102-150405"))

	sapSchedule := state.ObjAsSapNfsVolumeSnapshotSchedule()

	// Build labels: schedule-managed labels merged with template labels
	labels := map[string]string{
		cloudresourcesv1beta1.LabelScheduleName:      schedule.GetName(),
		cloudresourcesv1beta1.LabelScheduleNamespace: schedule.GetNamespace(),
	}
	maps.Copy(labels, sapSchedule.Spec.Template.Labels)

	// Build annotations from template
	annotations := make(map[string]string)
	maps.Copy(annotations, sapSchedule.Spec.Template.Annotations)

	// Determine deleteAfterDays: override with MaxRetentionDays if set
	deleteAfterDays := sapSchedule.Spec.Template.Spec.DeleteAfterDays
	if schedule.GetMaxRetentionDays() > 0 {
		deleteAfterDays = schedule.GetMaxRetentionDays()
	}

	snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   schedule.GetNamespace(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
			SourceVolume: corev1.ObjectReference{
				Name:      state.Source.Name,
				Namespace: state.Source.Namespace,
			},
			DeleteAfterDays: deleteAfterDays,
		},
	}

	// Check if snapshot already exists before creating
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      snapshot.Name,
		Namespace: snapshot.Namespace,
	}, &cloudresourcesv1beta1.SapNfsVolumeSnapshot{})
	if err != nil && apierrors.IsNotFound(err) {
		err = state.Cluster().K8sClient().Create(ctx, snapshot)
	}

	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		logger.Error(err, "Error creating SapNfsVolumeSnapshot")
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg(fmt.Sprintf("Error creating SapNfsVolumeSnapshot: %s", err)).
			Run(ctx, state)
	}

	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetBackupIndex(index)
	schedule.SetBackupCount(len(state.Snapshots) + 1)
	schedule.SetLastCreateRun(&metav1.Time{Time: state.nextRunTime.UTC()})
	schedule.SetLastCreatedBackup(corev1.ObjectReference{
		Kind:      "SapNfsVolumeSnapshot",
		Namespace: snapshot.Namespace,
		Name:      snapshot.Name,
	})
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
