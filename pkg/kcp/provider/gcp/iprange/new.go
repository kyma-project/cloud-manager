package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return v2.New(stateFactory.(*generalStateFactory).v2StateFactory)(ctx, st)
	}
}

// NewAllocateIpRangeAction returns an Action that will populate state.ExistingCidrRanges
// with occupied cidr ranges so the allocation can pick a free slot.
func NewAllocateIpRangeAction(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		return v2.NewAllocateIpRangeAction(stateFactory.(*generalStateFactory).v2StateFactory)(ctx, state)
	}
}
