package vpcpeering

import (
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitRemoteVpcPeeringAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering.GetState() != pb.NetworkPeering_INACTIVE.String() &&
		state.remoteVpcPeering.GetState() != pb.NetworkPeering_ACTIVE.String() {
		logger.Info("GCP Remote VPC Peering is not ready yet, re-queueing with delay")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, nil
}
