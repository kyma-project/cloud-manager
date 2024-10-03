package iprange

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
)

// New returns an Action that will provision and deprovision resource in the cloud.
// Common post actions are executed after it in the common iprange flow
// so in the case of success it must return nil error as a signal of success.
// If it returns non-nil error then it will break the common iprange flow
// immediately so it must as well set the error conditions properly.
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		ipRangeState := st.(iprangetypes.State)
		if ipRangeState.Network() == nil {
			if composed.MarkedForDeletionPredicate(ctx, st) {
				// in deprovisioning flow, should not be called w/out network, but if it is
				// then do nothing since cloud resources are already deprovisioned
				return nil, ctx
			}
			// LOGICAL ERROR!!!
			// in provisioning the flow, MUST not enter here w/out network
			return composed.LogErrorAndReturn(
				errors.New("logical error"),
				"Azure IpRange flow called w/out network in provisioning flow",
				composed.StopAndForget,
				ctx,
			)
		}
		state, err := stateFactory.NewState(ctx, ipRangeState)
		if err != nil {
			return err, ctx
		}
		return composed.ComposeActions(
			"azureIpRange",
			namesDetermine,
			securityGroupLoad,
			subnetLoad,
			privateDnsZoneLoad,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureIpRangeDelete",
					privateVirtualNetworkLinkDelete,
					privateDnsZoneDelete,
					subnetDelete,
					securityGroupDelete,
				),
				composed.ComposeActions(
					"azureIpRangeCreate",
					securityGroupCreate,
					securityGroupWait,
					subnetCreate,
					subnetWait,
					privateDnsZoneCreate,
					privateDnsZoneWait,
					privateVirtualNetworkLinkCreate,
					privateVirtualNetworkLinkWait,
				),
			),
		)(ctx, state)
	}
}
