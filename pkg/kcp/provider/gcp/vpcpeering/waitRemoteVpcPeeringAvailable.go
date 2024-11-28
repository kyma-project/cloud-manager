package vpcpeering

import (
	"context"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitRemoteVpcPeeringAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering.GetState() != pb.NetworkPeering_INACTIVE.String() &&
		state.remoteVpcPeering.GetState() != pb.NetworkPeering_ACTIVE.String() {
		logger.Info("[KCP GCP VpcPeering waitRemoteVpcPeeringActive] GCP Remote VPC Peering is not ready yet, re-queueing with delay", "currentState", state.remoteVpcPeering.GetState())
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, nil
}
