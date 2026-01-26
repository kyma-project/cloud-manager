package v3

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// validateCidr validates and parses the CIDR notation from status.cidr.
// It extracts the IP address and prefix length, storing them in state for use
// in GCP API calls. Invalid CIDR results in Error condition.
func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	logger = logger.WithValues("ipRange", ipRange.Name)
	logger.Info("Validating CIDR")

	// Require CIDR to be set
	if len(ipRange.Status.Cidr) == 0 {
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonInvalidCidr,
				Message: "CIDR is required",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error updating IpRange status due to missing CIDR").
			Run(ctx, state)
	}

	// Parse CIDR notation (e.g., "10.0.0.0/24" -> addr="10.0.0.0", prefix=24)
	addr, prefix, err := util.CidrParseIPnPrefix(ipRange.Status.Cidr)
	if err != nil {
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonInvalidCidr,
				Message: err.Error(),
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error updating IpRange status due to invalid CIDR").
			Run(ctx, state)
	}

	// Store parsed values in state for GCP API calls
	state.ipAddress = addr
	state.prefix = prefix

	logger.Info("CIDR validated", "ipAddress", addr, "prefix", prefix)

	return nil, nil
}
