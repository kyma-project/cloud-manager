package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// copyCidrToStatus copies the CIDR from spec to status if not already set.
// This ensures the status.cidr is populated even before allocation happens.
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
