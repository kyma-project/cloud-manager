package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	obj := state.ObjAsGcpNfsVolume()
	logger := composed.LoggerFromCtx(ctx).WithValues("id", obj.Name)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if obj.Status.Id != "" {
		logger.Info("Field .status.id is already set")
		return nil, nil
	}

	if state.KcpNfsInstance == nil {
		logger.Info("KCP NfsInstance does not exist")
		return nil, nil
	}

	obj.Status.Id = state.KcpNfsInstance.Name
	logger.Info("Updating .status.id with KCP NfsInstace name")
	return composed.PatchStatus(obj).
		SuccessError(composed.StopWithRequeue).
		SuccessLogMsg("Updated .status.id on GcpNfsVolume").
		ErrorLogMessage("Failed to update .status.id on GcpNfsVolume").
		Run(ctx, state)
}
