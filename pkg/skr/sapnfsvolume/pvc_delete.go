package sapnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func pvcDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PVC == nil {
		return nil, ctx
	}

	err := state.Cluster().K8sClient().Delete(ctx, state.PVC)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error deleting SapNfsVolume PVC", composed.StopWithRequeue, ctx)
	}

	if err != nil {
		logger.Info("Deleted SapNfsVolume PVC")
	}

	return nil, ctx
}
