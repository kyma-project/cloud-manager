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

	sanitized := composed.SanitizeConditions(state.KcpVpcPeering.Status.Conditions)

	changed := composed.AnyConditionChanged(obj, sanitized...) ||
		obj.Status.State != state.KcpVpcPeering.Status.State

	if changed {
		obj.Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(sanitized...).
			ErrorLogMessage("Error updating SKR AzureVpcPeering status").
			SuccessLogMsg("Updated and forgot SKR AzureVpcPeering status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, nil
}
