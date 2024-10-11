package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsScope().DeletionTimestamp.IsZero() {
		// Scope is being deleted
		return nil, nil
	}

	added, err := composed.PatchObjAddFinalizer(ctx, cloudcontrolv1beta1.FinalizerName, state.kyma, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with added finalizer", composed.StopWithRequeue, ctx)
	}

	if added {
		logger.Info("Added finalizer to the Kyma CR")
	}

	return nil, ctx
}
