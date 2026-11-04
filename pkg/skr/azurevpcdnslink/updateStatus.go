package azurevpcdnslink

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVNetLink()

	if state.KcpAzureVNetLink == nil {
		// it's deleted
		return nil, ctx
	}

	sanitized := composed.SanitizeConditions(state.KcpAzureVNetLink.Status.Conditions)

	stateChanged := obj.Status.State != state.KcpAzureVNetLink.Status.State
	if stateChanged {
		obj.Status.State = state.KcpAzureVNetLink.Status.State
	}
	changed := composed.SyncConditions(obj, sanitized...) || stateChanged

	if !changed {
		return nil, ctx
	}

	return composed.UpdateStatus(obj).
		ErrorLogMessage("Error updating SKR AzureVpcDnsLink status").
		SuccessLogMsg("Updated and forgot SKR AzureVpcDnsLink status").
		SuccessErrorNil().
		Run(ctx, state)
}
