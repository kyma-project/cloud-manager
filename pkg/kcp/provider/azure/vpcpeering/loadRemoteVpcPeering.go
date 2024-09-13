package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func loadRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	clientId := azureconfig.AzureConfig.PeeringCreds.ClientId
	clientSecret := azureconfig.AzureConfig.PeeringCreds.ClientSecret
	tenantId := state.tenantId

	resource, err := azureutil.ParseResourceID(obj.Spec.VpcPeering.Azure.RemoteVnet)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error parsing remote virtual network peering ID", composed.StopAndForget, ctx)
	}

	subscriptionId := resource.Subscription

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed to create azure loadRemoteVpcPeering client", composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx)
	}

	virtualNetworkName := resource.ResourceName
	resourceGroupName := resource.ResourceGroup
	virtualNetworkPeeringName := obj.Spec.VpcPeering.Azure.RemotePeeringName

	peering, err := c.GetPeering(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	if azuremeta.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading remote VPC Peering", ctx)
	}

	logger = logger.WithValues("remoteId", ptr.Deref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	remoteId := ptr.Deref(peering.ID, "")

	if obj.Status.RemoteId == remoteId {
		return nil, ctx
	}

	obj.Status.RemoteId = remoteId

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error updating VpcPeering status after loading vpc peering connection").
		SuccessErrorNil().
		Run(ctx, state)
}
