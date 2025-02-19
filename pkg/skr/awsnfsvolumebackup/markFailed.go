package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func markFailed(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If deletion, continue.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	backup := state.ObjAsAwsNfsVolumeBackup()
	backupState := backup.Status.State

	//If not in error state, continue
	if backupState != v1beta1.StateError {
		return nil, ctx
	}

	//If this backup doesn't belong to a schedule, continue
	scheduleName, exists := backup.GetLabels()[v1beta1.LabelScheduleName]
	if !exists {
		return nil, ctx
	}

	//createdOn := backup.GetCreationTimestamp().Format(time.RFC3339)
	list := &v1beta1.AwsNfsVolumeBackupList{}

	//List subsequent backups for this schedule.
	err := state.Cluster().K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			v1beta1.LabelScheduleName:      scheduleName,
			v1beta1.LabelScheduleNamespace: backup.GetNamespace(),
		},
		client.InNamespace(backup.GetNamespace()),
	)

	if err != nil {

		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonBackupListFailed,
				Message: fmt.Sprintf("Error listing subsequent backup(s) : %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//If there are subsequent backups exist,
	//mark this backup object state as failed.
	for _, item := range list.Items {

		if item.CreationTimestamp.Time.After(backup.CreationTimestamp.Time) {
			backup.Status.State = v1beta1.StateFailed
			return composed.PatchStatus(backup).
				SuccessLogMsg("AwsNfsVolumeBackup status updated with Failed state. ").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
	}

	//continue if there are no subsequent backups exist
	return nil, ctx
}
