package iprange

import (
	"context"
	"slices"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// rangeWaitBlockAssociated verifies that the secondary CIDR block is actually
// present in the VPC's secondary CIDR block list before proceeding to create
// vSwitches inside it. AliCloud's AssociateVpcCidrBlock may not take effect
// immediately, so creating a vSwitch too early would fail.
func rangeWaitBlockAssociated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	cidr := state.ObjAsIpRange().Status.Cidr

	// Re-read the VPC attribute to get the current secondary CIDR blocks
	vpcAttr, err := state.client.DescribeVpcAttribute(ctx, state.vpcId)
	if err != nil {
		logger.Error(err, "Error loading AliCloud VPC attribute while waiting for CIDR block association")
		return composed.StopWithRequeue, ctx
	}

	state.secondaryCidrBlocks = vpcAttr.SecondaryCidrBlocks

	if !slices.Contains(state.secondaryCidrBlocks, cidr) {
		logger.Info("Waiting for secondary CIDR block to be associated to AliCloud VPC", "cidr", cidr)
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	return nil, ctx
}
