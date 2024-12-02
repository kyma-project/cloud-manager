package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

func peeringLocalLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var resourceGroup, networkName string
	ok := false

	if len(state.ObjAsVpcPeering().Status.Id) > 0 {
		resourceID, err := util.ParseResourceID(state.ObjAsVpcPeering().Status.Id)
		if err == nil {
			resourceGroup = resourceID.ResourceGroup
			networkName = resourceID.ResourceName
			ok = true
		}

	}

	if !ok && state.localNetworkId != nil {
		resourceGroup = state.localNetworkId.ResourceGroup
		networkName = state.localNetworkId.NetworkName()
		ok = true
	}

	if !ok {
		logger.Info("Local Azure Peering not loaded")
		return nil, nil
	}

	// params must be the same as in peeringLocalCreate()
	peering, err := state.localClient.GetPeering(
		ctx,
		resourceGroup,
		networkName,
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
