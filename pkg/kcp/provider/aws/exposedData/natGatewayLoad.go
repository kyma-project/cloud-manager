package exposedData

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"k8s.io/utils/ptr"
)

func natGatewayLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpc == nil {
		return nil, ctx
	}

	gwList, err := state.awsClient.DescribeNatGateway(ctx, ptr.Deref(state.vpc.VpcId, ""))
	if x := awserrorhandling.HandleError(ctx, err, state, "ExposedData on list AWS NatGateways", cloudcontrolv1beta1.ReasonUnknown, "Error loading NAT Gateway"); x != nil {
		return x, ctx
	}

	for _, gw := range gwList {
		state.natGayteways = append(state.natGayteways, &gw)
	}

	return nil, ctx
}
