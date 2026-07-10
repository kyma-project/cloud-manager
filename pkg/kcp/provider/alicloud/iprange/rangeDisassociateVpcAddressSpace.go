package iprange

import (
	"context"
	"slices"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func rangeDisassociateVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	cidr := state.ObjAsIpRange().Status.Cidr
	if cidr == "" {
		return nil, ctx
	}

	if !slices.Contains(state.secondaryCidrBlocks, cidr) {
		return nil, ctx
	}

	logger.Info("Disassociating secondary CIDR block from AliCloud VPC", "vpcId", state.vpcId, "cidr", cidr)

	err := state.client.UnassociateVpcCidrBlock(ctx, state.vpcId, cidr)
	if err != nil {
		logger.Error(err, "Error disassociating secondary CIDR block from AliCloud VPC")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
