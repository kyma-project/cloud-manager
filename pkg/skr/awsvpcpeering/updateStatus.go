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

	sanitized := composed.SanitizeConditions(state.KcpVpcPeering.Status.Conditions)

	changed := composed.SyncConditions(state.ObjAsAwsVpcPeering(), sanitized...) ||
		state.ObjAsAwsVpcPeering().Status.State != state.KcpVpcPeering.Status.State

	if changed {
		state.ObjAsAwsVpcPeering().Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(state.ObjAsAwsVpcPeering()).
			SetExclusiveConditions(sanitized...).
			ErrorLogMessage("Error updating SKR AwsVpcPeering status").
			SuccessLogMsg("Updated and forgot SKR AwsVpcPeering status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, nil
}
