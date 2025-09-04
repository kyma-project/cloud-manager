package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.subnet == nil {
		return nil, ctx
	}

	subnet := state.ObjAsGcpSubnet()

	actualCidr := ptr.Deref(state.subnet.IpCidrRange, "")

	if subnet.Status.Cidr == actualCidr ||
		subnet.Spec.Cidr == "" && len(subnet.Status.Cidr) > 0 {
		return nil, ctx
	}

	subnet.Status.Cidr = actualCidr

	return composed.UpdateStatus(subnet).
		SuccessErrorNil().
		SuccessError(composed.StopWithRequeue).
		SuccessLogMsg("Subnet status updated with provisioned CIDR").
		ErrorLogMessage("Failed to update Subnet status to set provisioned CIDR").
		Run(ctx, state)
}
