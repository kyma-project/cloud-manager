package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func syncAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP Address")

	return nil, nil
}
