package v1

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventCidrEdit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.ObjAsIpRange().Status.Cidr) == 0 {
		// will be saved from next splitRangeByZones
		state.ObjAsIpRange().Status.Cidr = state.ObjAsIpRange().Spec.Cidr
	}

	if state.ObjAsIpRange().Spec.Cidr != state.ObjAsIpRange().Status.Cidr {
		readyCondition := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
		if readyCondition != nil && readyCondition.Status == metav1.ConditionTrue {
			meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
			meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCidrCanNotChange,
				Message: "Cidr can not change in Ready condition",
			})
			err := state.UpdateObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err, "Error updating IpRange status after Cidr change in ready condition", composed.StopWithRequeue, ctx)
			}
			return composed.StopAndForget, nil
		}
	}

	return nil, nil
}
