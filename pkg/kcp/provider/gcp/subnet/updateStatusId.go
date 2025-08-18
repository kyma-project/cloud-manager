package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	subnet := state.ObjAsGcpSubnet()

	if subnet.Status.Id != "" { // already set
		return nil, ctx
	}

	subnet.Status.Id = *state.subnet.Name

	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Subnet success .status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
