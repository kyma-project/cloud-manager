package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
)

// New returns an Action that will provision and deprovision resource in the cloud.
// Common post actions are executed after it in the common iprange flow
// so in the case of success it must return nil error as a signal of success.
// If it returns non-nil error then it will break the common iprange flow
// immediately so it must as well set the error conditions properly.
//
// TODO Phase 5: Replace this v2 wrapper with clean action composition
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		// Create v2 state factory wrapper for backward compatibility
		v2Factory := newV2StateFactoryAdapter(stateFactory)
		return v2.New(v2Factory)(ctx, st)
	}
}

// NewAllocateIpRangeAction returns an Action that will populate state.ExistingCidrRanges
// with occupied cidr ranges so the allocation can pick a free slot.
//
// TODO Phase 5: Replace this v2 wrapper with clean action composition
func NewAllocateIpRangeAction(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		// Create v2 state factory wrapper for backward compatibility
		v2Factory := newV2StateFactoryAdapter(stateFactory)
		return v2.NewAllocateIpRangeAction(v2Factory)(ctx, state)
	}
}

// v2StateFactoryAdapter adapts the new StateFactory to v2.StateFactory interface
// This is temporary until Phase 5 when we remove v2/ directory entirely
type v2StateFactoryAdapter struct {
	newStateFactory StateFactory
}

func newV2StateFactoryAdapter(sf StateFactory) v2.StateFactory {
	return &v2StateFactoryAdapter{newStateFactory: sf}
}

func (a *v2StateFactoryAdapter) NewState(ctx context.Context, ipRangeState iprangetypes.State) (*v2.State, error) {
	// Use the new state factory to create GCP state
	gcpState, err := a.newStateFactory.NewState(ctx, ipRangeState)
	if err != nil {
		return nil, err
	}

	// Convert to v2.State by creating a new v2 state with the same underlying data
	// v2.State embeds iprangetypes.State, and our new State also embeds it
	return v2.NewStateFromGcpState(gcpState), nil
}
