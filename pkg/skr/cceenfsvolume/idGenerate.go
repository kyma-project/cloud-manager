package cceenfsvolume

import (
	"context"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func idGenerate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsCceeNfsVolume().Status.Id != "" {
		return nil, ctx
	}

	state.ObjAsCceeNfsVolume().Status.Id = uuid.NewString()

	err := composed.PatchObjStatus(ctx, state.ObjAsCceeNfsVolume(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching CceeNfsVolume with status.id", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
