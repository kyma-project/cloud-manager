package gcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	nfsVolume := state.ObjAsGcpNfsVolume()

	//If deleting, continue with next steps.
	deleting := composed.IsMarkedForDeletion(state.Obj())
	if deleting {
		return nil, nil
	}

	logger.WithValues("GcpNfsVolume :", nfsVolume.Name).Info("Checking States")

	//If state is not empty, continue
	if nfsVolume.Status.State != "" {
		return nil, nil
	}

	//Set the state to processing
	nfsVolume.Status.State = cloudresourcesv1beta1.StateProcessing
	return composed.PatchStatus(nfsVolume).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
