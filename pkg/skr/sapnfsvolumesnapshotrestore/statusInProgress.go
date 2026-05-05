package sapnfsvolumesnapshotrestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusInProgress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	if restore.Status.State == cloudresourcesv1beta1.JobStateInProgress {
		return nil, ctx
	}

	restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
	return composed.PatchStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeProcessing,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreInProgress,
			Message: "Restore operation in progress",
		}).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
