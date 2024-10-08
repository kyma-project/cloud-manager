package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func skrDeactivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.moduleState != util.KymaModuleStateNotPresent {
		return nil, nil
	}
	if state.ObjAsScope() == nil || state.ObjAsScope().GetName() == "" {
		return nil, nil
	}

	logger.Info("Stopping SKR and deleting Scope")

	state.activeSkrCollection.RemoveScope(state.ObjAsScope())

	finalizerRemoved := controllerutil.RemoveFinalizer(state.Obj(), cloudcontrolv1beta1.FinalizerName)
	if finalizerRemoved {
		if err := state.UpdateObj(ctx); err != nil {
			return composed.LogErrorAndReturn(err, "Error updating Scope after finalizer removed", composed.StopWithRequeue, ctx)
		}
	}

	if err := state.Cluster().K8sClient().Delete(ctx, state.Obj()); err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Scope", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
