package iprange

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
)

func New(sf StateFactory) composed.Action {
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
				"SAP IpRange flow called w/out network in provisioning flow",
				composed.StopAndForget,
				ctx,
			)
		}

		state, err := sf.NewState(ctx, ipRangeState)
		if err != nil {
			return fmt.Errorf("error creating sap iprange state: %w", err), ctx
		}

		return composed.ComposeActionsNoName(
			networkLoad,
			routerLoad,
			subnetLoad,
			routerSubnetLoad,
			composed.If(
				composed.NotMarkedForDeletionPredicate,
				// create/update
				subnetCreate,
				routerSubnetAdd,
				statusPatch,
			),
			composed.If(
				composed.MarkedForDeletionPredicate,
				// delete
				routerSubnetRemove,
				subnetDelete,
			),
		)(ctx, state)
	}
}
