package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		runtimeState := state.(runtimetypes.State)
		ctx, state, err := sf.NewState(ctx, runtimeState)
		if err != nil {
			return err, ctx
		}

		return composed.ComposeActionsNoName(
			composed.Noop,
		)(ctx, state)
	}
}
