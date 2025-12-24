package v3

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// identifyPeeringIpRanges builds the list of IP ranges to include in PSA peering.
// When creating/updating PSA connection, we need to include all PSA-purpose IP ranges
// that exist in the VPC, not just the current one.
// Note: This action is only called in the create-update branch, never during deletion.
func identifyPeeringIpRanges(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	ipRange := state.ObjAsIpRange()

	// If the address object doesn't exist in GCP, continue
	if state.address == nil {
		return nil, nil
	}

	// If Service Connection doesn't exist in GCP
	if state.serviceConnection == nil {
		// Add this IP address to be included while creating Service Connection
		if state.address.Name != nil {
			state.peeringIpRanges = []string{*state.address.Name}
		}
		return nil, nil
	}

	logger.WithValues("ipRange", ipRange.Name).Info("Loading IpRanges List")

	// Get GCP scope specific values
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	// List all global addresses in the VPC
	list, err := state.computeClient.ListGlobalAddresses(ctx, project, vpc)
	if err != nil {
		logger.Error(err, "Error listing Global Addresses from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error listing Global Addresses from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Load the address names into a map (only PSA purpose addresses)
	tmpMap := map[string]struct{}{}
	for _, addr := range list {
		if isPsaPurpose(addr) {
			if addr.Name != nil {
				tmpMap[*addr.Name] = struct{}{}
			}
		}
	}

	// Build peeringIpRanges directly from all PSA addresses that exist in GCP
	// Order doesn't matter for GCP PSA connection validation
	state.peeringIpRanges = make([]string, 0, len(tmpMap))
	for name := range tmpMap {
		state.peeringIpRanges = append(state.peeringIpRanges, name)
	}

	return nil, nil
}

// isPsaPurpose checks if an address has PSA purpose.
// Handles pointer field in computepb.Address.
func isPsaPurpose(addr *computepb.Address) bool {
	if addr.Purpose == nil {
		return false
	}
	return *addr.Purpose == string(v1beta1.GcpPurposePSA)
}
