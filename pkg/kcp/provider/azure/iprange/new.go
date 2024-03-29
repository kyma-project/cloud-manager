package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := stateFactory.NewState(st.(focal.State))
		return composed.ComposeActions(
			"gcpIpRange")(ctx, state)
	}
}
