package kyma

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrDeactivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.
		WithValues("skrKymaName", state.ObjAsKyma().GetName()).
		Info("Stopping SKR")

	state.activeSkrCollection.RemoveKyma(state.ObjAsKyma())

	return nil, ctx
}
