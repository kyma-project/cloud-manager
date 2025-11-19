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

	if state.remoteOperation != nil && state.remoteOperation.GetError() != nil {
		return nil, ctx
	}

	if state.localVpcPeering != nil && vpcPeering.Status.Id == "" {
		statusChanged = true
		logger.Info("[KCP GCP VpcPeering setPeeringStatusIds] setting Local connection id status " + vpcPeering.Status.Id)
		vpcPeering.Status.Id = ptr.Deref(state.localVpcPeering.Name, "")
	}

	if state.remoteVpcPeering != nil && vpcPeering.Status.RemoteId == "" {
		statusChanged = true
		logger.Info("[KCP GCP VpcPeering setPeeringStatusIds] setting remote connection id status " + vpcPeering.Status.RemoteId)
		vpcPeering.Status.RemoteId = ptr.Deref(state.remoteVpcPeering.Name, "")
	}
	if statusChanged {
		logger.Info("[KCP GCP VpcPeering setPeeringStatusIds] attempting to patch id status")
		return composed.PatchStatus(vpcPeering).
			SuccessLogMsg("[KCP GCP VpcPeering setPeeringStatusIds] Successfully patched status ids").
			ErrorLogMessage("[KCP GCP VpcPeering setPeeringStatusIds] Error patching status").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
