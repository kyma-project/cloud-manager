package awsnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	added := controllerutil.AddFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	if !added {
		// finalizer already added
		return nil, nil
	}

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving AwsNfsVolume after finalizer added", composed.StopWithRequeue, nil)
	}
	err = state.Reloader.ReloadObjKindOneKey(ctx, state.Obj(), client.ObjectKeyFromObject(state.Obj()))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error reloading AwsIpRange after finaler added", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
