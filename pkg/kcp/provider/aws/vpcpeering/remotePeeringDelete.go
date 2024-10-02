package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func remotePeeringDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		return nil, nil
	}

	if state.remoteVpcPeering == nil {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	logger.Info("Deleting remote VpcPeering")

	err := state.remoteClient.DeleteVpcPeeringConnection(ctx, state.remoteVpcPeering.VpcPeeringConnectionId)

	if err != nil {

		if awsmeta.IsNotFound(err) {
			logger.Info("VpcPeeringConnection not found")
			return nil, nil
		}

		return awsmeta.LogErrorAndReturn(err, "Error deleting vpc peering", ctx)
	}

	logger.Info("Remote VpcPeering deleted")

	return nil, nil
}
