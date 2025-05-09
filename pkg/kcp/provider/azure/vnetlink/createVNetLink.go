package vnetlink

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink != nil {
		return nil, ctx
	}

	kymaNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	kymaResourceGroup := state.Scope().Spec.Scope.Azure.VpcNetwork
	vnetId := azureutil.NewVirtualNetworkResourceId(state.Scope().Spec.Scope.Azure.SubscriptionId,
		kymaResourceGroup, kymaNetworkName).String()

	err := state.remoteClient.CreateVirtualNetworkLink(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName,
		state.ObjAsAzureVNetLink().Spec.RemoteVirtualPrivateLinkName,
		vnetId)

	if err == nil {
		logger.Info("VirtualNetworkLink created")
		return nil, ctx
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on creating VirtualNetworkLink",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	return azuremeta.LogErrorAndReturn(err, "Error creating VirtualNetworkLink", ctx)
}
