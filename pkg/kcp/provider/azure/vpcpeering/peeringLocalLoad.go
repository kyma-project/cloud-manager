package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func peeringLocalLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// params must be the same as in peeringLocalCreate()
	peering, err := state.localClient.GetPeering(
		ctx,
		state.localNetworkId.ResourceGroup,
		state.localNetworkId.NetworkName(),
		state.ObjAsVpcPeering().GetLocalPeeringName(),
	)
	if azuremeta.IsNotFound(err) {
		logger.Info("Local Azure Peering not found for KCP VpcPeering")
		return nil, nil
	}
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading VPC Peering", ctx)
	}

	logger = logger.WithValues("localPeeringId", ptr.Deref(peering.ID, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.localPeering = peering

	logger.Info("Azure local Peering loaded for KCP VpcPeering")

	return nil, ctx
}
