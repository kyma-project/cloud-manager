package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.localVpcId == "" {
		return nil, ctx
	}

	tables, err := state.client.DescribeRouteTables(ctx, state.localVpcId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading local route tables", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.routeTables = tables

	return nil, ctx
}
