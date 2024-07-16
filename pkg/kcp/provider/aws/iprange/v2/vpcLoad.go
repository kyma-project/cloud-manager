package v2

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"k8s.io/utils/ptr"
)

func vpcLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsIpRange().Status.VpcId) == 0 {
		return nil, nil
	}

	vpc, err := state.awsClient.DescribeVpc(ctx, state.ObjAsIpRange().Status.VpcId)
	if vpc == nil {
		err = fmt.Errorf("vpc %s not found", state.ObjAsIpRange().Status.VpcId)
	}
	if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on VPC not found",
		cloudcontrolv1beta1.ReasonVpcNotFound, "VPC not found"); x != nil {
		return x, nil
	}

	state.vpc = vpc

	logger = logger.WithValues("vpcId", ptr.Deref(state.vpc.VpcId, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
