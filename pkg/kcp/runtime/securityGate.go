package runtime

import (
	"context"
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

var defaultSecurityGate = &securityGate{
	actualStateRuntimes:      map[string]bool{},
	actualStateSubscriptions: map[string]bool{},
}

type securityActualState struct {
	enabledOnRuntime      *bool
	enabledOnSubscription *bool
}

type securityGate struct {
	m sync.Mutex

	actualStateRuntimes      map[string]bool
	actualStateSubscriptions map[string]bool
}

func (g *securityGate) getActualState(runtimeId, subscriptionId string) securityActualState {
	g.m.Lock()
	defer g.m.Unlock()

	var result securityActualState
	if v, ok := g.actualStateRuntimes[runtimeId]; ok {
		result.enabledOnRuntime = new(v)
	}
	if v, ok := g.actualStateSubscriptions[subscriptionId]; ok {
		result.enabledOnSubscription = new(v)
	}
	return result
}

func (g *securityGate) markSuccess(ds runtimetypes.SecurityDesiredState) {
	g.m.Lock()
	defer g.m.Unlock()

	g.actualStateRuntimes[ds.RuntimeId()] = ds.EnabledOnRuntime()
	g.actualStateSubscriptions[ds.SubscriptionId()] = ds.EnabledOnSubscription()
}

func (g *securityGate) shouldRun(ds runtimetypes.SecurityDesiredState) bool {
	as := g.getActualState(ds.RuntimeId(), ds.SubscriptionId())

	if as.enabledOnRuntime == nil || as.enabledOnSubscription == nil {
		// no track so far, this is the first run in this cloud-manager instance, go do reconcile cloud
		return true
	}
	if *as.enabledOnRuntime != ds.EnabledOnRuntime() || *as.enabledOnSubscription != ds.EnabledOnSubscription() {
		// desired and actual state are different, go do reconcile cloud
		return true
	}

	// actual and desired state are equal, no need to reconcile
	return false
}

type SecurityDesiredStatable interface {
	SecurityDesiredState() runtimetypes.SecurityDesiredState
}

func (g *securityGate) ShouldRunPredicate(_ context.Context, st composed.State) bool {
	state := st.(SecurityDesiredStatable)
	ds := state.SecurityDesiredState()
	if ds == nil {
		// shouldn't happen if reconciler flow is ok, but just in case
		// runtime or subscription resources not loaded -> unknown ids -> can not determine, can not reconcile
		return false
	}
	return g.shouldRun(ds)
}

var _ composed.Predicate = (*securityGate)(nil).ShouldRunPredicate
