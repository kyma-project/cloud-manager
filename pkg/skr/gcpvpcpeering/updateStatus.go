package gcpvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpVpcPeering == nil {
		// it's deleted
		return nil, nil
	}

	if composed.SyncConditions(state.ObjAsGcpVpcPeering(), *state.KcpVpcPeering.Conditions()...) ||
		state.ObjAsGcpVpcPeering().Status.State != state.KcpVpcPeering.Status.State {
		state.ObjAsGcpVpcPeering().Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(state.ObjAsGcpVpcPeering()).
			SetExclusiveConditions(state.KcpVpcPeering.Status.Conditions[0]).
			ErrorLogMessage(state.KcpVpcPeering.Status.Conditions[0].Message).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
