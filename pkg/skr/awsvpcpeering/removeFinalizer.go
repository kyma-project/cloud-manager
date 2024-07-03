package awsvpcpeering

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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

	// KCP VpcPeering does not exist, remove the finalizer so SKR AwsVpcPeering is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)

	err := state.UpdateObj(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AzureVpcPeering after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye AwsVpcPeering
	return composed.StopAndForget, nil
}
