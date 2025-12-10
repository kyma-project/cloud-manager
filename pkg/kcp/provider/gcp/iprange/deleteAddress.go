package iprange

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// deleteAddress deletes the GCP global address resource.
// This is called during IpRange deletion to clean up the address reservation.
func deleteAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if address doesn't exist
	if state.address == nil {
		logger.Info("Address does not exist, nothing to delete")
		return nil, ctx
	}

	ipRange := state.ObjAsIpRange()
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project

	// Get the address name from the loaded address resource
	addressName := ""
	if state.address.Name != nil {
		addressName = *state.address.Name
	}

	logger = logger.WithValues(
		"ipRange", ipRange.Name,
		"addressName", addressName,
	)

	logger.Info("Deleting GCP Address")

	// Delete the global address
	operationName, err := state.computeClient.DeleteIpRange(ctx, project, addressName)

	if err != nil {
		logger.Error(err, "Error deleting Address from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error deleting Address: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Updated condition for failed Address deletion").
			Run(ctx, state)
	}

	// Store operation identifier for tracking
	ipRange.Status.OpIdentifier = operationName

	logger.Info("Address deletion initiated", "operation", operationName)

	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
		Run(ctx, state)
}
