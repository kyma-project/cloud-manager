package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsNfsInstance().Status.State != cloudcontrolv1beta1.ReadyState {
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ReadyState
		changed = true
	}

	if len(state.ObjAsNfsInstance().Status.Conditions) != 1 {
		changed = true
	}

	cond := meta.FindStatusCondition(*state.ObjAsNfsInstance().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ReasonReady {
		changed = true
	}

	if !changed {
		return nil, nil
	}

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Ready",
		}).
		ErrorLogMessage("Error patching CCEE NfsInstance status after setting ready condition").
		Run(ctx, state)
}
