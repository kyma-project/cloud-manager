package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/pointer"
)

func loadRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	clientId := azureconfig.AzureConfig.ClientId
	clientSecret := azureconfig.AzureConfig.ClientSecret
	tenantId := state.tenantId

	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	resource, err := util.ParseResourceID(obj.Status.RemoteId)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error parsing remote virtual network peering ID", nil)
	}

	subscriptionId := resource.Subscription

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	peering, err := c.Get(ctx, resource.ResourceGroup, resource.ResourceName, resource.SubResourceName)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading remote VPC Peering", nil)
	}

	logger = logger.WithValues("remoteId", pointer.StringDeref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	return nil, ctx
}
