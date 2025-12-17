package v3

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// createAddress creates a new GCP global address for the IpRange.
// It uses the allocated CIDR (state.ipAddress and state.prefix) to create
// the address resource in GCP with PSC (Private Service Connect) configuration.
func createAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Skip if address already exists
	if state.address != nil {
		return nil, ctx
	}

	ipRange := state.ObjAsIpRange()

	// Validate that CIDR has been parsed (defensive programming)
	if state.ipAddress == "" || state.prefix == 0 {
		logger.Error(fmt.Errorf("missing CIDR data"), "Cannot create address without valid CIDR")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonInvalidCidr,
				Message: "CIDR must be validated before creating address",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error: CIDR not validated before address creation").
			Run(ctx, state)
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork
	name := GetIpRangeName(ipRange.GetName())

	logger = logger.WithValues(
		"ipRange", ipRange.Name,
		"addressName", name,
		"ipAddress", state.ipAddress,
		"prefix", state.prefix,
	)

	logger.Info("Creating GCP Address")

	// Create the global address with PSC configuration
	operationName, err := state.computeClient.CreatePscIpRange(
		ctx,
		project,
		vpc,
		name,
		"Kyma cloud-manager IP Range",
		state.ipAddress,
		int64(state.prefix),
	)

	if err != nil {
		logger.Error(err, "Error creating Address in GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error creating Address: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Updated condition for failed Address creation").
			Run(ctx, state)
	}

	// Store operation identifier for tracking
	ipRange.Status.OpIdentifier = operationName

	logger.Info("Address creation initiated", "operation", operationName)

	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
		Run(ctx, state)
}
