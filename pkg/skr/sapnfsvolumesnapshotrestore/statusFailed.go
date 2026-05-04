package sapnfsvolumesnapshotrestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusFailed(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	restore.Status.State = cloudresourcesv1beta1.JobStateFailed
	return composed.PatchStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
			Message: "Restore operation failed",
		}).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
