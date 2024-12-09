package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.ObjAsNuke().Status.State = "Completed"

	return composed.PatchStatus(state.ObjAsNuke()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeReady,
			Message: "All orphans deleted",
		}).
		Run(ctx, state)
}
