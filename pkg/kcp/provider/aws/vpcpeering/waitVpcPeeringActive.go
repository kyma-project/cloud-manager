package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering.Status.Code != ec2types.VpcPeeringConnectionStateReasonCodeActive {
		logger.Info("Waiting for peering Connected state",
			"Id", *state.vpcPeering.VpcPeeringConnectionId,
			"PeeringStatus", state.vpcPeering.Status.Code)
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, nil
}
