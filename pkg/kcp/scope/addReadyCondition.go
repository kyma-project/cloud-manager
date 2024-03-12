package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	scope := state.ObjAsScope()
	return composed.UpdateStatus(scope).
		SetCondition(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Ready",
		}).
		ErrorLogMessage("Error updating scope status with ready condition").
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}).
		Run(ctx, state)
}
