package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsNuke().Status.State == "Completed" {
		composed.LoggerFromCtx(ctx).Info("Nuke is completed, short-circuiting into StopAndForget")
		return composed.StopAndForget, nil
	}

	return nil, ctx
}
