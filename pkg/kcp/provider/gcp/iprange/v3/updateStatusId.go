package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	ipRange := state.ObjAsIpRange()

	if ipRange.Status.Id != "" { // already set
		return nil, nil
	}

	ipRange.Status.Id = *state.subnet.Name

	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success .status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
