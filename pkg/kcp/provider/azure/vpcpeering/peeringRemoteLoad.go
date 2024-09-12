package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func peeringRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// params must be the same as in peeringRemoteCreate()
	peering, err := state.remoteClient.GetPeering(
		ctx,
		state.remoteNetworkId.ResourceGroup,
		state.remoteNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
	)
	if azuremeta.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading remote VPC Peering", ctx)
	}

	logger = logger.WithValues("remotePeeringId", ptr.Deref(peering.ID, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	return nil, ctx
}
