package v2

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	ipRange := state.ObjAsIpRange()
	logger := composed.LoggerFromCtx(ctx).WithValues("ipRange", ipRange.Name)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.address == nil {
		logger.Info("Address is not created yet")
		return nil, nil
	}

	if ipRange.Status.Id != "" {
		logger.Info("Field .status.id is already set")
		return nil, nil
	}

	ipRange.Status.Id = state.address.Name
	logger.Info("Updating .status.id with gcp iprange identifier")
	return composed.PatchStatus(ipRange).
		SuccessError(composed.StopWithRequeue).
		SuccessLogMsg("Updated .status.id on IpRange").
		ErrorLogMessage("Failed to update .status.id on IpRange").
		Run(ctx, state)
}
