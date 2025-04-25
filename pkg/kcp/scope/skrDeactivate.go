package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrDeactivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.activeSkrCollection.RemoveScope(ctx, state.ObjAsScope())

	return nil, ctx
}
