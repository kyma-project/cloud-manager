package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func scopeSave(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.Obj().GetResourceVersion() != "" {
		// Existing scope, update it
		err := state.Cluster().K8sClient().Update(ctx, state.ObjAsScope())
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating scope", composed.StopWithRequeue, ctx)
		}
		composed.LoggerFromCtx(ctx).Info("Scope updated")
	} else {
		// New scope, create it
		err := state.Cluster().K8sClient().Create(ctx, state.ObjAsScope())
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating scope", composed.StopWithRequeue, ctx)
		}
		composed.LoggerFromCtx(ctx).Info("Scope created")
	}

	return nil, ctx
}
