package sapnfsvolumesnapshot

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	if snapshot.Status.State == cloudresourcesv1beta1.StateDeleting {
		return nil, ctx
	}

	snapshot.Status.State = cloudresourcesv1beta1.StateDeleting
	return composed.PatchStatus(snapshot).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: "Deleting snapshot",
		}).
		SuccessError(nil).
		Run(ctx, state)
}
