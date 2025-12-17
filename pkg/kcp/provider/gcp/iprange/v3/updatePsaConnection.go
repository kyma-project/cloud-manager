package v3

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"google.golang.org/api/servicenetworking/v1"
)

// updatePsaConnection updates an existing Private Service Access (PSA) connection.
// This is called when the list of reserved IP ranges changes (adding/removing IpRanges).
// GCP uses PATCH operation to update the connection with the new list of ranges.
// This action is idempotent - it only updates when the ranges actually differ.
func updatePsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if no PSA connection exists
	if state.serviceConnection == nil {
		logger.Info("No PSA connection to update")
		return nil, ctx
	}

	// Check if the connection already has the desired IP ranges (idempotency)
	if state.DoesConnectionMatchPeeringRanges() {
		logger.Info("PSA connection already has correct IP ranges, skipping update")
		return nil, ctx
	}

	ipRange := state.ObjAsIpRange()
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	logger = logger.WithValues(
		"ipRange", ipRange.Name,
		"project", project,
		"vpc", vpc,
		"ipRanges", state.peeringIpRanges,
		"existingRanges", state.serviceConnection.ReservedPeeringRanges,
	)

	logger.Info("Updating GCP PSA Connection")

	// Set state to indicate PSA connection sync in progress
	ipRange.Status.State = gcpclient.SyncPsaConnection

	var operation interface{}
	var err error

	// If no IP ranges left, delete the connection
	if len(state.peeringIpRanges) == 0 {
		logger.Info("No IP ranges left, deleting PSA connection")
		operation, err = state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)
	} else {
		// Update the connection with new IP ranges
		operation, err = state.serviceNetworkingClient.PatchServiceConnection(
			ctx,
			project,
			vpc,
			state.peeringIpRanges,
		)
	}

	if err != nil {
		logger.Error(err, "Error updating PSA Connection in GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error updating PSA Connection: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Updated condition for failed PSA Connection update").
			Run(ctx, state)
	}

	// Store operation identifier for tracking
	if operation != nil {
		if op, ok := operation.(*servicenetworking.Operation); ok && op != nil {
			ipRange.Status.OpIdentifier = op.Name
			logger.Info("PSA Connection update initiated", "operation", op.Name)
		}
	}

	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
