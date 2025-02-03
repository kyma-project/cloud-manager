package azurevpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpVpcPeering != nil {
		// KCP VpcPeering is not yet deleted
		return nil, nil
	}

	logger.Info("Removing AzureVpcPeering finalizer")

	// KCP VpcPeering does not exist, remove the finalizer so SKR AzureVpcPeering is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)

	err := state.UpdateObj(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AzureVpcPeering after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye AzureVpcPeering
	return composed.StopAndForget, nil
}
