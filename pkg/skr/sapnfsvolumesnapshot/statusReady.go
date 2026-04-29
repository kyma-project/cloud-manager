package sapnfsvolumesnapshot

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	if state.snapshot == nil {
		return nil, ctx
	}

	snapshot.Status.State = cloudresourcesv1beta1.StateReady
	snapshot.Status.SizeGb = state.snapshot.Size

	return composed.PatchStatus(snapshot).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonReady,
			Message: "Snapshot is ready",
		}).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
