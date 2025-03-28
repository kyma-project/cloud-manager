package kyma

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrActivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.
		WithValues("skrKymaName", state.ObjAsKyma().GetName()).
		Info("Activating SKR")

	state.activeSkrCollection.AddKyma(state.ObjAsKyma())

	return nil, ctx
}
