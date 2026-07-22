package iprange

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
)

// New returns an Action that will provision and deprovision vSwitches in AliCloud
// inside a secondary CIDR block associated to the Gardener VPC.
func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		ipRangeState := st.(iprangetypes.State)
		if ipRangeState.Network() == nil {
			if composed.MarkedForDeletionPredicate(ctx, st) {
				return nil, ctx
			}
			return composed.LogErrorAndReturn(
				errors.New("logical error"),
				"AliCloud IpRange flow called w/out network in provisioning flow",
				composed.StopAndForget,
				ctx,
			)
		}

		state, err := sf.NewState(ctx, ipRangeState)
		if err != nil {
			return fmt.Errorf("error creating alicloud iprange state: %w", err), ctx
		}

		return composed.ComposeActionsNoName(
			vpcLoad,
			vSwitchLoad,
			composed.If(
				composed.NotMarkedForDeletionPredicate,
				// create
				rangeSplitByZones,
				rangeCheckPrimaryOverlap,
				rangeCheckVSwitchOverlap,
				rangeExtendVpcAddressSpace,
				rangeWaitBlockAssociated,
				vSwitchCreate,
				vSwitchWait,
				statusPatch,
			),
			composed.If(
				composed.MarkedForDeletionPredicate,
				// delete
				vSwitchDelete,
				rangeDisassociateVpcAddressSpace,
			),
		)(ctx, state)
	}
}
