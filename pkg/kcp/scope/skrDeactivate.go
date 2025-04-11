package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrDeactivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Stopping SKR")

	state.activeSkrCollection.RemoveScope(state.ObjAsScope())

	return nil, ctx
}
