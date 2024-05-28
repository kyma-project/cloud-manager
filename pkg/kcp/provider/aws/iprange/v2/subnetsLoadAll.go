package v2

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/pointer"
)

func subnetsLoadAll(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	subnetList, err := state.awsClient.DescribeSubnets(ctx, pointer.StringDeref(state.vpc.VpcId, ""))
	if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on load subnets",
		cloudcontrolv1beta1.ReasonUnknown, "Error loading AWS VPC subnets"); x != nil {
		return x, nil
	}
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading subnets", ctx)
	}

	state.allSubnets = subnetList

	return nil, nil
}
