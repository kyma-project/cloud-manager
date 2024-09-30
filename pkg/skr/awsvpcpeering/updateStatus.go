package awsvpcpeering

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

	changed := false

	if composed.AnyConditionChanged(state.ObjAsAwsVpcPeering(), *state.KcpVpcPeering.Conditions()...) {
		changed = true
	}

	if state.ObjAsAwsVpcPeering().Status.State != state.KcpVpcPeering.Status.State {
		changed = true
	}

	if changed {
		state.ObjAsAwsVpcPeering().Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(state.ObjAsAwsVpcPeering()).
			SetExclusiveConditions(*state.KcpVpcPeering.Conditions()...).
			ErrorLogMessage("Error updating SKR AwsVpcPeering status").
			SuccessLogMsg("Updated and forgot SKR AwsVpcPeering status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, nil
}
