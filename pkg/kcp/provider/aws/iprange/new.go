package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/v1"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return v1.New(stateFactory.(*generealStateFactory).v1StateFactory)(ctx, st)
	}
}
