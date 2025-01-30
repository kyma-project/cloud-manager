package cceenfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func finalizerAdd(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	_, err := composed.PatchObjAddFinalizer(ctx, cloudresourcesv1beta1.Finalizer, state.ObjAsCceeNfsVolume(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error adding finalizer to CceeNfsVolume", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
