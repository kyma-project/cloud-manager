package v3

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// prepareAllocateIpRange sets up the existing CIDR ranges from the GCP scope
// before calling the shared allocateIpRange logic.
// This is the GCP-specific setup that prepares state.existingCidrRanges with
// the network CIDR ranges (nodes, pods, services) from the Shoot/Kyma runtime.
func prepareAllocateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	// Validate that we have network information
	if len(state.Scope().Spec.Scope.Gcp.Network.Nodes) == 0 {
		logger.Error(nil, "SKR nodes range is unknown in scope")
		ipRange.Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
				Message: "Error due to unknown SKR nodes range",
			}).
			ErrorLogMessage("Failed patching KCP IpRange status with error due to unknown SKR nodes range").
			SuccessLogMsg("Forgetting KCP IpRange in error state due to unknown SKR nodes range").
			Run(ctx, st)
	}

	// Set existing CIDR ranges from the GCP scope
	// These are the network ranges that are already in use and should be avoided
	state.SetExistingCidrRanges([]string{
		state.Scope().Spec.Scope.Gcp.Network.Nodes,
		state.Scope().Spec.Scope.Gcp.Network.Pods,
		state.Scope().Spec.Scope.Gcp.Network.Services,
	})

	logger.Info("Prepared existing CIDR ranges for allocation",
		"nodes", state.Scope().Spec.Scope.Gcp.Network.Nodes,
		"pods", state.Scope().Spec.Scope.Gcp.Network.Pods,
		"services", state.Scope().Spec.Scope.Gcp.Network.Services,
	)

	return nil, ctx
}
