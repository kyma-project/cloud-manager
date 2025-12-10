package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// updateStatusId updates the status.id field with the GCP address name.
// This happens after the address is successfully created in GCP.
// The ID is used to track the GCP resource and for reference by other resources.
func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	ipRange := state.ObjAsIpRange()
	logger := composed.LoggerFromCtx(ctx).WithValues("ipRange", ipRange.Name)

	// Skip if marked for deletion
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// Skip if address not yet created
	if state.address == nil {
		logger.Info("Address is not created yet")
		return nil, nil
	}

	// Skip if ID already set
	if ipRange.Status.Id != "" {
		logger.Info("Field .status.id is already set")
		return nil, nil
	}

	// Extract name from address (handling pointer field)
	addressName := ""
	if state.address.Name != nil {
		addressName = *state.address.Name
	}

	// Update status with GCP resource identifier
	ipRange.Status.Id = addressName
	logger.Info("Updating .status.id with GCP iprange identifier", "id", addressName)

	return composed.PatchStatus(ipRange).
		SuccessError(composed.StopWithRequeue).
		SuccessLogMsg("Updated .status.id on IpRange").
		ErrorLogMessage("Failed to update .status.id on IpRange").
		Run(ctx, state)
}
