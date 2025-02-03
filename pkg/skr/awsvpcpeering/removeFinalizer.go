package awsvpcpeering

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

	logger.Info("Removing AwsVpcPeering finalizer")

	// KCP VpcPeering does not exist, remove the finalizer so SKR AwsVpcPeering is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)

	err := state.UpdateObj(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AwsVpcPeering after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye AwsVpcPeering
	return composed.StopAndForget, nil
}
