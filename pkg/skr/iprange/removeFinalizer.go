package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.KcpIpRange != nil {
		// KCP IpRange is not yet deleted
		return nil, nil
	}

	logger.Info("Removing SKR IpRange finalizer")

	// KCP IpRange does not exist, remove the finalizer so SKR IpRange is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR IpRange after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye SKR IpRange
	return composed.StopAndForget, nil
}
