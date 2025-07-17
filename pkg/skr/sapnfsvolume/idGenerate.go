package sapnfsvolume

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func idGenerate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsSapNfsVolume().Status.Id != "" {
		return nil, ctx
	}

	state.ObjAsSapNfsVolume().Status.Id = uuid.NewString()

	if len(state.ObjAsSapNfsVolume().Status.State) == 0 {
		state.ObjAsSapNfsVolume().Status.State = cloudresourcesv1beta1.StateCreating
	}

	err := composed.PatchObjStatus(ctx, state.ObjAsSapNfsVolume(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching SapNfsVolume with status.id", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
