package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsScope().Status.State != cloudcontrolv1beta1.ReadyState {
		state.ObjAsScope().Status.State = cloudcontrolv1beta1.ReadyState
		changed = true
	}

	if len(state.ObjAsScope().Status.Conditions) != 1 {
		changed = true
	}

	cond := meta.FindStatusCondition(*state.ObjAsScope().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ReasonReady || cond.Message != cloudcontrolv1beta1.ReasonReady {
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsScope()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: cloudcontrolv1beta1.ReasonReady,
		}).
		ErrorLogMessage("Error patching scope status with ready condition").
		SuccessErrorNil().
		Run(ctx, state)
}
