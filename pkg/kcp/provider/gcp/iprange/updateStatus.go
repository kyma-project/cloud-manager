package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// updateStatus updates the IpRange status to Ready when provisioning is complete.
// This action should be called after:
// - Address is created in GCP
// - PSA connection is created/updated (if PSA purpose)
// - All async operations have completed
// - Status fields (id, cidr) are populated
// This action is idempotent - it only updates when status is not already Ready.
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	// Check if already Ready (idempotency)
	readyCondition := meta.FindStatusCondition(ipRange.Status.Conditions, v1beta1.ConditionTypeReady)
	if readyCondition != nil && readyCondition.Status == metav1.ConditionTrue {
		logger.Info("IpRange already has Ready status, skipping update")
		return nil, nil
	}

	// Skip if there's an ongoing operation
	if ipRange.Status.OpIdentifier != "" {
		logger.Info("Skipping status update, operation in progress")
		return nil, nil
	}

	// Verify address exists in GCP
	if state.address == nil {
		logger.Info("Address not yet created, skipping Ready status")
		return nil, nil
	}

	// Check if PSA connection is required and exists
	gcpOptions := ipRange.Spec.Options.Gcp
	if gcpOptions == nil || gcpOptions.Purpose == v1beta1.GcpPurposePSA {
		// For PSA purpose, verify PSA connection exists and includes this range
		if state.serviceConnection == nil {
			logger.Info("PSA connection not yet created, skipping Ready status")
			return nil, nil
		}

		// Verify this IP range is included in the PSA connection
		if state.DoesConnectionIncludeRange() < 0 {
			logger.Info("IP range not yet included in PSA connection, skipping Ready status")
			return nil, nil
		}
	}

	// All prerequisites met, set Ready condition
	logger.Info("IpRange is fully provisioned, setting Ready status")

	return composed.PatchStatus(ipRange).
		SetExclusiveConditions(metav1.Condition{
			Type:    v1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.ReasonReady,
			Message: "IpRange provisioned in GCP",
		}).
		SuccessError(composed.StopAndForget).
		SuccessLogMsg("IpRange is Ready").
		Run(ctx, state)
}
