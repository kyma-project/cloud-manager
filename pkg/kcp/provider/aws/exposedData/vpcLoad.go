package exposedData

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
)

func vpcLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	vpcList, err := state.awsClient.DescribeVpcs(ctx, state.vpcName)
	if x := awserrorhandling.HandleError(ctx, err, state, "ExposedData on list AWS VPCs", cloudcontrolv1beta1.ReasonUnknown, "Error loading VPC"); x != nil {
		return x, nil
	}

	if len(vpcList) > 0 {
		state.vpc = &vpcList[0]
	}

	return nil, ctx
}
