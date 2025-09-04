package feature

import (
	"context"
)

const vpcPeeringSyncFlagName = "vpcPeeringSync"

var VpcPeeringSync = &vpcPeeringSync{}

type vpcPeeringSync struct{}

func (f *vpcPeeringSync) Value(ctx context.Context) bool {
	v := provider.BoolVariation(ctx, vpcPeeringSyncFlagName, false)
	return v
}
