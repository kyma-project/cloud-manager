package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func loadSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	subnetList, err := state.client.DescribeSubnets(ctx, pointer.StringDeref(state.vpc.VpcId, ""))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading subnets", composed.StopWithRequeue, ctx)
	}

	state.allSubnets = subnetList

	return nil, nil
}
