package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func fixStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Fix Status.Id")

	//If the object is marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If the Status.Id already present OR address object is nil, continue
	if ipRange.Status.Id != "" || state.address == nil {
		return nil, nil
	}

	ipRange.Status.Id = state.address.Address
	return composed.UpdateStatus(ipRange).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
