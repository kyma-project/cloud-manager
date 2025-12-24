package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// copyCidrToStatus initializes the status.cidr field from spec.cidr if not already set.
// Since CIDR is immutable (enforced by webhook validation), this action only runs once
// during the initial reconciliation to populate status.cidr before GCP allocation.
func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	ipRange := state.ObjAsIpRange()

	// Skip if already set
	if len(ipRange.Status.Cidr) > 0 {
		return nil, nil
	}

	// Copy from spec to status
	ipRange.Status.Cidr = ipRange.Spec.Cidr

	return composed.PatchStatus(ipRange).
		SuccessErrorNil().
		Run(ctx, state)
}
