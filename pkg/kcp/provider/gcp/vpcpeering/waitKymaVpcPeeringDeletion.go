package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitKymaVpcPeeringDeletion(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering != nil {
		logger.Info("GCP Kyma VPC Peering is not deleted yet, re-queueing with delay")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, nil
}
