package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func setPeeringStatusIds(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	vpcPeering := state.ObjAsVpcPeering()
	statusChanged := false

	if state.remotePeeringOperation != nil && state.remotePeeringOperation.GetError() != nil {
		return nil, ctx
	}

	if state.localVpcPeering != nil && vpcPeering.Status.Id == "" {
		statusChanged = true
		vpcPeering.Status.Id = ptr.Deref(state.localVpcPeering.Name, "")
		logger.Info("setting Local connection id status ", "localId", vpcPeering.Status.Id)
	}

	if state.remoteVpcPeering != nil && vpcPeering.Status.RemoteId == "" {
		statusChanged = true
		vpcPeering.Status.RemoteId = ptr.Deref(state.remoteVpcPeering.Name, "")
		logger.Info("setting remote connection id status ", "remoteId", vpcPeering.Status.RemoteId)
	}
	if statusChanged {
		logger.Info("attempting to patch id status")
		return composed.PatchStatus(vpcPeering).
			SuccessLogMsg("Successfully patched status ids").
			ErrorLogMessage("Error patching status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
