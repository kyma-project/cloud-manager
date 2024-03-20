package iprange

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/pointer"
	"time"
)

func disassociateVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
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

		return composed.StopWithRequeueDelay(300 * time.Millisecond), ctx
	}

	if outcome == util.Ignore {
		return nil, ctx
	}

	logger.Info("Disassociating VPC Cidr block")

	err := state.client.DisassociateVpcCidrBlockInput(ctx, pointer.StringDeref(theBlock.AssociationId, ""))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error disassociating VPC Cidr block", composed.StopWithRequeueDelay(300*time.Millisecond), ctx)
	}

	return composed.StopWithRequeueDelay(300 * time.Millisecond), ctx
}
