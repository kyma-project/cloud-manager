package kyma

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrActivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.activeSkrCollection.AddKyma(ctx, state.ObjAsKyma())

	return nil, ctx
}
