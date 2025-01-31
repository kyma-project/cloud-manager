package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func scopeDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.moduleState != util.KymaModuleStateNotPresent {
		return nil, nil
	}
	if state.ObjAsScope() == nil || state.ObjAsScope().GetName() == "" {
		return nil, nil
	}

	logger.Info("Deleting Scope")

	if _, err := state.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook); err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Scope after finalizer removed", composed.StopWithRequeue, ctx)
	}

	if err := state.Cluster().K8sClient().Delete(ctx, state.Obj()); err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Scope", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
