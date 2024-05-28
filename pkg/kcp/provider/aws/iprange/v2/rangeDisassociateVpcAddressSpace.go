package v2

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/pointer"
)

func rangeDisassociateVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var theBlock *ec2Types.VpcCidrBlockAssociation
	for _, cidrBlock := range state.vpc.CidrBlockAssociationSet {
		if pointer.StringDeref(cidrBlock.CidrBlock, "") == state.ObjAsIpRange().Spec.Cidr {
			theBlock = &cidrBlock
		}
	}

	if theBlock == nil {
		return nil, nil
	}

	if theBlock.CidrBlockState == nil {
		logger.Info("VPC Cidr block without state")
		return nil, nil
	}

	logger = logger.WithValues("cidrBlockState", theBlock.CidrBlockState.State)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	actMap := util.NewDelayActIgnoreBuilder[ec2Types.VpcCidrBlockStateCode](util.Ignore).
		Delay(ec2Types.VpcCidrBlockStateCodeAssociating).
		Act(ec2Types.VpcCidrBlockStateCodeAssociated).
		Build()

	outcome := actMap.Case(theBlock.CidrBlockState.State)

	if outcome == util.Delay {
		logger.Info("Waiting for VPC Cidr block state")

		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	if outcome == util.Ignore {
		return nil, ctx
	}

	logger.Info("Disassociating VPC Cidr block")

	err := state.awsClient.DisassociateVpcCidrBlockInput(ctx, pointer.StringDeref(theBlock.AssociationId, ""))
	if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange after DisassociateVpcCidrBlock",
		cloudcontrolv1beta1.ReasonUnknown, "Failed deleting VPC CIDR address block"); x != nil {
		return x, nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
