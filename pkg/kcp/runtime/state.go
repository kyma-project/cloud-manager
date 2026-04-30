package runtime

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

var _ runtimetypes.State = &State{}

type State struct {
	composed.State

	subscription *cloudcontrolv1beta1.Subscription

	allRuntimesInSubscription map[string]bool

	// securityServiceEnabledOnSubscription true if any runtime in the subscription has security enabled
	// indicates that security scanning services producing findings has to
	securityServiceEnabledOnSubscription bool
	securityDataSourceEnabledOnRuntime   bool

	securityCooldown *securityCooldown
}

func newState(baseState composed.State) *State {
	return &State{
		State:            baseState,
		securityCooldown: newSecurityCooldown(time.Minute * 5),
	}
}

func (s *State) ObjAsRuntime() *infrastructuremanagerv1.Runtime {
	return s.Obj().(*infrastructuremanagerv1.Runtime)
}

func (s *State) Subscription() *cloudcontrolv1beta1.Subscription {
	return s.subscription
}

func (s *State) SecurityServiceEnabledOnSubscription() bool {
	return s.securityServiceEnabledOnSubscription
}

func (s *State) SecurityDataSourceEnabledOnRuntime() bool {
	return s.securityDataSourceEnabledOnRuntime
}

func (s *State) SecurityServiceEnabledOnSubscriptionPredicate(_ context.Context, st composed.State) bool {
	return s.SecurityServiceEnabledOnSubscription()
}

func (s *State) SecurityDataSourceEnabledOnRuntimePredicate(_ context.Context, st composed.State) bool {
	return s.SecurityDataSourceEnabledOnRuntime()
}
