package v3

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// createPsaConnection creates a new Private Service Access (PSA) connection.
// PSA connections enable VPC networks to connect to Google-managed services.
// The connection is created with the list of reserved IP ranges (including this IpRange).
func createPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if PSA connection already exists
	if state.serviceConnection != nil {
		logger.Info("PSA connection already exists")
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
	)

	logger.Info("Creating GCP PSA Connection")

	// Create the PSA connection with the reserved IP ranges
	operation, err := state.serviceNetworkingClient.CreateServiceConnection(
		ctx,
		project,
		vpc,
		state.peeringIpRanges,
	)

	if err != nil {
		logger.Error(err, "Error creating PSA Connection in GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error creating PSA Connection: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Updated condition for failed PSA Connection creation").
			Run(ctx, state)
	}

	// Store operation identifier for tracking
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		logger.Info("PSA Connection creation initiated", "operation", operation.Name)
	}

	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
		Run(ctx, state)
}
