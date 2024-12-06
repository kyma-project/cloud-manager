package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcPeering.Status.Code == ec2types.VpcPeeringConnectionStateReasonCodeActive {
		return nil, nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
