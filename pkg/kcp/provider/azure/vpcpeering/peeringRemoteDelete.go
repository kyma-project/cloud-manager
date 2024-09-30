package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func peeringRemoteDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	lll := logger.WithValues("vpcPeeringName", state.ObjAsVpcPeering().Spec.Details.PeeringName)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		return nil, nil
	}

	if len(state.ObjAsVpcPeering().Status.RemoteId) == 0 {
		lll.Info("Remote VpcPeering deleted before Azure peering is created")
		return nil, nil
	}

	// params must be the same as in peeringRemoteCreate()
	err := state.remoteClient.DeletePeering(
		ctx,
		state.remoteNetworkId.ResourceGroup,
		state.remoteNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
	)

	lll = lll.WithValues("vpcPeeringId", state.ObjAsVpcPeering().Status.RemoteId)
	lll.Info("Deleting VpcPeering")

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting vpc peering", composed.LoggerIntoCtx(ctx, lll))
	}

	lll.Info("Remote VpcPeering deleted")

	return nil, nil
}
