package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func peeringLocalLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var resourceGroup, networkName string
	ok := false

	if len(state.ObjAsVpcPeering().Status.Id) > 0 {
		resourceID, err := azureutil.ParseResourceID(state.ObjAsVpcPeering().Status.Id)
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
		logger.Info("Local VPC peering not loaded")
		return nil, ctx
	}

	// params must be the same as in peeringLocalCreate()
	peering, err := state.localClient.GetPeering(
		ctx,
		resourceGroup,
		networkName,
		state.ObjAsVpcPeering().GetLocalPeeringName(),
	)

	if err == nil {
		logger = logger.WithValues("localPeeringId", ptr.Deref(peering.ID, ""))
		ctx = composed.LoggerIntoCtx(ctx, logger)

		state.localPeering = peering

		logger.Info("Local VPC peering loaded")

		return nil, ctx
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on loading local VPC peering",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	if azuremeta.IsNotFound(err) {
		logger.Info("Local VPC peering not found")
		return nil, ctx
	}

	return azuremeta.LogErrorAndReturn(err, "Error loading local VPC peering", ctx)
}
