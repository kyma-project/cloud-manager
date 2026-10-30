package vpcpeering

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" {
		return nil, ctx
	}

	// If the connection is already gone, nothing to do.
	if state.vpcPeering != nil && tea.StringValue(state.vpcPeering.Status) == alicloudvpcpeeringclient.StatusDeleted {
		return nil, ctx
	}

	// While the connection is still transitioning, wait.
	if state.vpcPeering != nil && tea.StringValue(state.vpcPeering.Status) == alicloudvpcpeeringclient.StatusDeleting {
		logger.Info("AliCloud VpcPeerConnection is still Deleting, requeueing")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	if err := state.client.DeleteVpcPeerConnection(ctx, instanceId); err != nil {
		if alicloudvpcpeeringclient.IsNotFound(err) {
			return nil, ctx
		}
		if alicloudvpcpeeringclient.IsRetryable(err) {
			// IncorrectStatus fires when the connection is already transitioning; safe to retry.
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}
		return composed.LogErrorAndReturn(err, "Error deleting AliCloud VpcPeerConnection", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	// Deletion is async — requeue to let loadVpcPeerConnection confirm it is gone.
	logger.Info("AliCloud VpcPeerConnection delete issued, requeueing to confirm")
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
