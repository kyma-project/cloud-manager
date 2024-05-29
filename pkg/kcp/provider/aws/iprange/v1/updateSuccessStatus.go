package v1

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateSuccessStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), cloudcontrolv1beta1.ConditionTypeError)
	meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ReasonReady,
		Message: "Additional IpRange(s) are provisioned",
	})
	state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ReadyState

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
