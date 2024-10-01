package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func removeKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kyma == nil {
		return nil, ctx
	}

	removed, err := composed.PatchObjRemoveFinalizer(ctx, cloudcontrolv1beta1.FinalizerName, state.kyma, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with removed finalizer", composed.StopWithRequeue, ctx)
	}

	if removed {
		logger.Info("Removed finalizer from the Kyma CR")
	}

	return nil, ctx
}
