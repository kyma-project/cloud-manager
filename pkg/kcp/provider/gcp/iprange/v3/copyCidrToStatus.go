package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.subnet == nil {
		return nil, nil
	}

	ipRange := state.ObjAsIpRange()

	actualCidr := ptr.Deref(state.subnet.IpCidrRange, "")

	if ipRange.Status.Cidr == actualCidr ||
		ipRange.Spec.Cidr == "" && len(ipRange.Status.Cidr) > 0 {
		return nil, nil
	}

	ipRange.Status.Cidr = actualCidr

	return composed.UpdateStatus(ipRange).
		SuccessErrorNil().
		SuccessError(composed.StopWithRequeue).
		SuccessLogMsg("IpRange status updated with provisioned CIDR").
		ErrorLogMessage("Failed to update IpRange status to set provisioned CIDR").
		Run(ctx, state)
}
