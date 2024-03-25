package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	isRemoved := controllerutil.RemoveFinalizer(state.ObjAsScope(), cloudcontrolv1beta1.FinalizerName)

	if !isRemoved {
		return nil, nil
	}

	logger.Info("Removing finalizer from the Kyma CR")

	err := state.Cluster().K8sClient().Update(ctx, state.kyma)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with removed finalizer", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
