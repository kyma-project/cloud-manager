package criprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.KcpIpRange != nil {
		// KCP IpRange is not yet deleted
		return nil, nil
	}

	// KCP IpRange does not exist, remove the finalizer so SKR IpRange is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR IpRange after finalizer remove", composed.StopWithRequeue, nil)
	}

	// bye, bye SKR IpRange
	return composed.StopAndForget, nil
}
