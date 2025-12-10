package iprange

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
// When deleting, we exclude the current IP range from the list.
func identifyPeeringIpRanges(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	ipRange := state.ObjAsIpRange()

	// Check whether the object is being deleted
	deleting := composed.IsMarkedForDeletion(ipRange)

	// If the address object doesn't exist in GCP, continue
	if state.address == nil {
		return nil, nil
	}

	// If Service Connection doesn't exist in GCP
	if state.serviceConnection == nil {
		// Add this IP address to be included while creating Service Connection
		if !deleting {
			// Handle pointer field
			if state.address.Name != nil {
				state.peeringIpRanges = []string{*state.address.Name}
			}
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

	// Reset the peeringIpRanges slice
	state.peeringIpRanges = []string{}

	// Iterate over the list of peering ranges, and include required ones
	for _, name := range state.serviceConnection.ReservedPeeringRanges {
		// If it is the name of current IP range, skip it (we'll add it later if not deleting)
		if state.address.Name != nil && name == *state.address.Name {
			continue
		}

		// If the IP range exists in GCP, include it
		if _, ok := tmpMap[name]; ok {
			state.peeringIpRanges = append(state.peeringIpRanges, name)
		}
	}

	// If not deleting, add the name of this IpRange
	if !deleting && state.address.Name != nil {
		state.peeringIpRanges = append(state.peeringIpRanges, *state.address.Name)
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
