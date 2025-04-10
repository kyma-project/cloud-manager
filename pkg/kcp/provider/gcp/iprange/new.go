package iprange

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// New returns an Action that will provision and deprovision resource in the cloud.
// Common post actions are executed after it in the common iprange flow
// so in the case of success it must return nil error as a signal of success.
// If it returns non-nil error then it will break the common iprange flow
// immediately so it must as well set the error conditions properly.
func New(stateFactory StateFactory) composed.Action {
	return composed.BuildSwitchAction(
		"typeSwitch",
		v2.New(stateFactory.(*generalStateFactory).v2StateFactory), // default
		composed.NewCase(TypeGlobalAddressPredicate, v2.New(stateFactory.(*generalStateFactory).v2StateFactory)),
		composed.NewCase(TypeSubnetPredicate, v3.New(stateFactory.(*generalStateFactory).v3StateFactory)),
	)
}

// NewAllocateIpRangeAction returns an Action that will populate state.ExistingCidrRanges
// with occupied cidr ranges so the allocation can pick a free slot.
func NewAllocateIpRangeAction(stateFactory StateFactory) composed.Action {
	return v2.NewAllocateIpRangeAction(stateFactory.(*generalStateFactory).v2StateFactory)
}
