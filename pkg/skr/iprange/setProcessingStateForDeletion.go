package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setProcessingStateForDeletion(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}
	if state.KcpIpRange == nil || composed.IsMarkedForDeletion(state.KcpIpRange) {
		return nil, nil // KCP IpRange is already marked for deletion, so it already passed Processing state
	}

	// Only update status if not already in Processing state to avoid unnecessary updates and conflicts
	if state.ObjAsIpRange().Status.State == cloudresourcesv1beta1.StateProcessing {
		return nil, nil
	}

	state.ObjAsIpRange().SetState(cloudresourcesv1beta1.StateProcessing)
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR IpRange status with Processing state", composed.StopWithRequeue, ctx)
	}
	return nil, nil
}
