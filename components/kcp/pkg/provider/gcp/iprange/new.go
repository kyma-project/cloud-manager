package iprange

import (
	"context"
	"github.com/kyma-project/cloud-resources-manager/components/kcp/pkg/common/actions/scope"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := stateFactory.NewState(st.(scope.State))
		return composed.ComposeActions(
			"gcpIpRange")(ctx, state)
	}
}
