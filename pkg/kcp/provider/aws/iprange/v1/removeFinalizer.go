package v1

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	isUpdated := controllerutil.RemoveFinalizer(state.ObjAsIpRange(), cloudcontrolv1beta1.FinalizerName)
	if !isUpdated {
		return nil, nil
	}

	logger.Info("Removing finalizer")

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP IpRange after finalizer removed", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
