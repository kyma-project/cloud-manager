package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

func peeringRemoteDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		return nil, nil
	}

	if state.remotePeering == nil {
		logger.Info("Azure remote peering not loaded, continuing")
		return nil, nil
	}

	resourceId, err := util.ParseResourceID(ptr.Deref(state.remotePeering.ID, ""))

	if err != nil {
		logger.Error(err, "Failed parsing remotePeering.ID while deleting remote peering")
		return nil, nil
	}
	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	// params must be the same as in peeringRemoteCreate()
	err = state.remoteClient.DeletePeering(
		ctx,
		resourceId.ResourceGroup,
		resourceId.ResourceName,
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
	)

	logger.Info("Deleting remote VpcPeering")

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting vpc peering", nil, ctx)
	}

	logger.Info("Remote VpcPeering deleted")

	return nil, nil
}
