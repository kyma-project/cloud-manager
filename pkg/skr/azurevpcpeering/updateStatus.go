package azurevpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureVpcPeering()

	if state.KcpVpcPeering == nil {
		// it's deleted
		return nil, nil
	}

	if composed.AnyConditionChanged(obj, *state.KcpVpcPeering.Conditions()...) || obj.Status.State != state.KcpVpcPeering.Status.State {
		obj.Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(*state.KcpVpcPeering.Conditions()...).
			ErrorLogMessage("Error updating SKR AzureVpcPeering status").
			SuccessLogMsg("Updated and forgot SKR AzureVpcPeering status").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, nil
}
