package v2

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func finalizerAdd(ctx context.Context, state composed.State) (error, context.Context) {
	// Object is being deleted, don't add finalizer
	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	modified, err := state.PatchObjAddFinalizer(ctx, cloudcontrolv1beta1.FinalizerName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP IpRage after finalizer added", composed.StopWithRequeue, ctx)
	}
	if modified {
		composed.LoggerFromCtx(ctx).Info("Finalizer added")
	}

	return nil, nil
}
