package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func requeueIfNeeded(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.hadErrors {
		return composed.StopWithRequeueDelay(rate.Default.Slow1s.When(state.Obj())), ctx
	}
	return nil, ctx
}
