package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func providerFromScopeToState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsObjLoaded(ctx, state) {
		return nil, ctx
	}

	state.provider = state.ObjAsScope().Spec.Provider

	return nil, ctx
}
