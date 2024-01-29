package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func updateState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	state.ObjAsIpRange().Status.State = state.curState

	if state.curState == v1beta1.ReadyState {
		meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), v1beta1.ConditionTypeError)
		state.AddReadyCondition(ctx, "IpRange provisioned in GCP.")
	}

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	if state.inSync && state.curState == v1beta1.ReadyState {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
