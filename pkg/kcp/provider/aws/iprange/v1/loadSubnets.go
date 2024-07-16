package v1

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	subnetList, err := state.client.DescribeSubnets(ctx, ptr.Deref(state.vpc.VpcId, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading subnets", ctx)
	}

	state.allSubnets = subnetList

	return nil, nil
}
