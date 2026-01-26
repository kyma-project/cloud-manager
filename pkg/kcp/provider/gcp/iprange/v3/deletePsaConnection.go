package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deletePsaConnection deletes the Private Service Access connection from GCP.
// This is called during IpRange deletion when this is the last IpRange using the PSA connection.
// Note: PSA connections are shared across all IpRanges in a VPC, so we only delete
// when no more IpRanges are using it.
func deletePsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if no PSA connection exists
	if state.serviceConnection == nil {
		logger.Info("No PSA connection to delete")
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
	)

	logger.Info("Deleting GCP PSA Connection")

	// Set state to indicate PSA connection deletion in progress
	ipRange.Status.State = gcpclient.DeletePsaConnection

	// Delete the PSA connection
	operation, err := state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)

	if err != nil {
		logger.Error(err, "Error deleting PSA Connection from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Updated condition for failed PSA Connection deletion").
			Run(ctx, state)
	}

	// Store operation identifier for tracking
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		logger.Info("PSA Connection deletion initiated", "operation", operation.Name)
	}

	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
