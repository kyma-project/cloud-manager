package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	v1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v1"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		autoCidrAllocationEnabled := feature.IpRangeAutomaticCidrAllocation.Value(ctx)
		logger := composed.LoggerFromCtx(ctx)
		logger.WithValues("autoCidrAllocationEnabled", autoCidrAllocationEnabled).Info("Auto CIDR allocation flag")
		if !autoCidrAllocationEnabled {
			return v1.New(stateFactory.(*generalStateFactory).v1StateFactory)(ctx, st)
		} else {
			return v2.New(stateFactory.(*generalStateFactory).v2StateFactory)(ctx, st)
		}
	}
}
