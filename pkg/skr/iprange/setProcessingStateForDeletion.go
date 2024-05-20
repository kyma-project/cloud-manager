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
	if composed.IsMarkedForDeletion(state.KcpIpRange) {
		return nil, nil // KCP IpRange is already marked for deletion, so it already passed Processing state
	} else {
		state.ObjAsIpRange().SetState(cloudresourcesv1beta1.StateProcessing)
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			// No reason to halt the flow if we can't set the processing state here as it will go to "deleting" or "error" state soon
			return composed.LogErrorAndReturn(err, "Error updating SKR IpRange status with Processing state", nil, ctx)
		}
		return nil, nil
	}
}
