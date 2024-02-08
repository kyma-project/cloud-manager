package iprange

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	ipRange := state.ObjAsIpRange()
	prevState := ipRange.Status.State
	ipRange.Status.State = state.curState

	if state.curState == v1beta1.ReadyState {
		ipRange.Status.Cidr = ipRange.Spec.Cidr
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonReady,
				Message: "IpRange provisioned in GCP.",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	} else if prevState != state.curState {
		return composed.UpdateStatus(ipRange).SuccessError(composed.StopWithRequeue).Run(ctx, state)
	}

	return nil, nil
}
