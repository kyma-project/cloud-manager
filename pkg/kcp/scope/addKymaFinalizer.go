package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func addKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsScope().DeletionTimestamp.IsZero() {
		// Scope is being deleted
		return nil, nil
	}

	added := controllerutil.AddFinalizer(state.kyma, cloudcontrolv1beta1.FinalizerName)

	if !added {
		// finalizer already added on Kyma CR
		return nil, nil
	}

	logger.Info("Adding finalizer to the Kyma CR")

	err := state.Cluster().K8sClient().Update(ctx, state.kyma)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with added finalizer", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	return composed.StopWithRequeueDelay(time.Second), nil
}
