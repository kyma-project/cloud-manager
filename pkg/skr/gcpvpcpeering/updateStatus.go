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

	sanitized := composed.SanitizeConditions(state.KcpVpcPeering.Status.Conditions)

	if composed.SyncConditions(state.ObjAsGcpVpcPeering(), sanitized...) ||
		state.ObjAsGcpVpcPeering().Status.State != state.KcpVpcPeering.Status.State {
		state.ObjAsGcpVpcPeering().Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(state.ObjAsGcpVpcPeering()).
			SetExclusiveConditions(sanitized...).
			ErrorLogMessage("Error updating SKR GcpVpcPeering status").
			SuccessLogMsg("Updated and forgot SKR GcpVpcPeering status").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
