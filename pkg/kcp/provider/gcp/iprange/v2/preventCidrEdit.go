package v2

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventCidrEdit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.ObjAsIpRange().Spec.Cidr) == 0 {
		return nil, nil
	}
	if len(state.ObjAsIpRange().Status.Cidr) == 0 {
		return nil, nil
	}

	if state.ObjAsIpRange().Spec.Cidr != state.ObjAsIpRange().Status.Cidr {
		readyCondition := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
		if readyCondition != nil && readyCondition.Status == metav1.ConditionTrue {
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  "True",
					Reason:  cloudcontrolv1beta1.ReasonCidrCanNotChange,
					Message: "Cidr can not change in Ready condition",
				}).
				ErrorLogMessage("Error patching IpRange status after Cidr change in ready condition").
				SuccessLogMsg("Forgetting KCP IpRange with changed CIDR in Ready state").
				Run(ctx, st)
		}
	}

	return nil, nil
}
