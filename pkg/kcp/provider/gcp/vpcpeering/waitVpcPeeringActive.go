package vpcpeering

import (
	"context"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if (state.localPeeringOperation != nil && state.localPeeringOperation.GetError() != nil) || (state.remotePeeringOperation != nil && state.remotePeeringOperation.GetError() != nil) {
		return nil, ctx
	}

	if state.localVpcPeering.GetState() != pb.NetworkPeering_ACTIVE.String() && state.remoteVpcPeering.GetState() != pb.NetworkPeering_ACTIVE.String() {
		logger.Info("GCP VPC Peering is not ready yet, re-queueing with delay", "currentState", state.remoteVpcPeering.GetState())
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	return nil, nil
}
