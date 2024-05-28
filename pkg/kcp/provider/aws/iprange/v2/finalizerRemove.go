package v2

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func finalizerRemove(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	modified, err := state.PatchObjRemoveFinalizer(ctx, cloudcontrolv1beta1.FinalizerName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP IpRange after finalizer removed", composed.StopWithRequeue, ctx)
	}
	if modified {
		logger.Info("Finalizer removed")

	}

	return nil, nil
}
