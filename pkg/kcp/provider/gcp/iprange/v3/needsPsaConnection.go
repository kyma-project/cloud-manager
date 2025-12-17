package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// needsPsaConnection determines if a PSA connection is needed for this IpRange.
// PSA connection is required when:
// - The address exists in GCP
// - The purpose is PSA (or not explicitly set to non-PSA)
func needsPsaConnection(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	ipRange := state.ObjAsIpRange()

	// No PSA connection if address doesn't exist yet
	if state.address == nil {
		return false
	}

	// Check if purpose is PSA (default is PSA if not specified)
	gcpOptions := ipRange.Spec.Options.Gcp
	if gcpOptions == nil {
		return true // Default to PSA
	}

	return gcpOptions.Purpose == v1beta1.GcpPurposePSA
}
