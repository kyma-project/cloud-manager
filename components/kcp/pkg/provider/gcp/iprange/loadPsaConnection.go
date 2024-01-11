package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func loadPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()

	//If this IPRange is not for PSA, no processing is needed here.
	if ipRange.Spec.Options.Gcp == nil ||
		ipRange.Spec.Options.Gcp.Purpose != v1beta1.GcpPurposePSA {
		return nil, nil
	}

	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP PSA Connection")

	return nil, nil
}
