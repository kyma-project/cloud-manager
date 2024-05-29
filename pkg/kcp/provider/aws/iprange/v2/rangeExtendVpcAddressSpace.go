package v2

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"k8s.io/utils/pointer"
)

func rangeExtendVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.associatedCidrBlock != nil {
		return nil, nil
	}

	logger.Info("Associating vpc cidr block")

	block, err := state.awsClient.AssociateVpcCidrBlock(ctx, pointer.StringDeref(state.vpc.VpcId, ""), state.ObjAsIpRange().Status.Cidr)
	if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on associate vpc cidr block",
		cloudcontrolv1beta1.ReasonFailedExtendingVpcAddressSpace, "Failed extending vpc address space"); x != nil {
		return x, nil
	}

	state.ObjAsIpRange().Status.AddressSpaceId = pointer.StringDeref(block.AssociationId, "")

	return composed.PatchStatus(state.ObjAsIpRange()).
		SuccessErrorNil().
		Run(ctx, state)
}
