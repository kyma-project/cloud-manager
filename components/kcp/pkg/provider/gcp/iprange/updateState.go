package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func updateState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	state.ObjAsIpRange().Status.State = state.curState

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	if state.inSync && state.curState == v1beta1.ReadyState {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
