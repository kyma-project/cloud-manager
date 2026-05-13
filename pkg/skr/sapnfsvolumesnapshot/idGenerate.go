package sapnfsvolumesnapshot

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func idGenerate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	if snapshot.Status.Id != "" {
		return nil, ctx
	}

	snapshot.Status.Id = uuid.NewString()
	snapshot.Status.State = cloudresourcesv1beta1.StateCreating

	err := composed.PatchObjStatus(ctx, snapshot, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching snapshot with status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, ctx
}
