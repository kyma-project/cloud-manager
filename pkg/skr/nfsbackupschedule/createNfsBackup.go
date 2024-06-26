package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func createNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
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

	//If the creation for the nextRunTime is already done, return
	if schedule.Status.LastCreateRun != nil && !schedule.Status.LastCreateRun.IsZero() &&
		state.nextRunTime.Unix() == schedule.Status.LastCreateRun.Time.Unix() {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info(fmt.Sprintf("Creation already completed for %s ", state.nextRunTime))
		return nil, nil
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Creating File Backup")

	//Get Backup details (name and index).
	prefix := schedule.Spec.Prefix
	if prefix == "" {
		prefix = schedule.Name
	}
	index := schedule.Status.BackupIndex + 1
	name := fmt.Sprintf("%s-%d-%s", prefix, index, state.nextRunTime.UTC().Format("20060102-150405"))

	//Create Backup object
	backup, err := getProviderSpecificBackupObject(state, name)
	if err == nil {
		err = state.Cluster().K8sClient().Create(ctx, backup)
	}

	if err != nil {
		schedule.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error updating File backup object in GCP :%s", err)).
			Run(ctx, state)
	}

	//Update the status of the schedule with the latest backup details
	schedule.Status.State = cloudresourcesv1beta1.JobStateActive
	schedule.Status.BackupIndex = index
	schedule.Status.LastCreateRun = &metav1.Time{Time: state.nextRunTime.UTC()}
	schedule.Status.Backups = append(schedule.Status.Backups, corev1.ObjectReference{
		Kind:      backup.GetObjectKind().GroupVersionKind().Kind,
		Namespace: backup.GetNamespace(),
		Name:      backup.GetName(),
	})
	return composed.UpdateStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}

func getProviderSpecificBackupObject(state *State, name string) (client.Object, error) {
	schedule := state.ObjAsNfsBackupSchedule()

	//Construct a common part of NfsBackup Object
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: schedule.Namespace,
		Labels: map[string]string{
			"nfs-backup-schedule":    schedule.Name,
			"nfs-backup-schedule-ns": schedule.Namespace,
		},
	}

	//Construct provider specific NfsBackup Object
	switch state.Scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		return &cloudresourcesv1beta1.GcpNfsVolumeBackup{
			ObjectMeta: objectMeta,
			Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
				Location: schedule.Spec.Location,
				Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
					Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
						Name:      schedule.Spec.NfsVolumeRef.Name,
						Namespace: schedule.Spec.NfsVolumeRef.Namespace,
					},
				},
			},
		}, nil
	case cloudcontrolv1beta1.ProviderAws:
		return &cloudresourcesv1beta1.AwsNfsVolumeBackup{
			ObjectMeta: objectMeta,
			Spec: cloudresourcesv1beta1.AwsNfsVolumeBackupSpec{
				Source: cloudresourcesv1beta1.AwsNfsVolumeBackupSource{
					Volume: cloudresourcesv1beta1.VolumeRef{
						Name:      schedule.Spec.NfsVolumeRef.Name,
						Namespace: schedule.Spec.NfsVolumeRef.Namespace,
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}

}
