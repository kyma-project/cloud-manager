package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	clientId := azureconfig.AzureConfig.PeeringCreds.ClientId
	clientSecret := azureconfig.AzureConfig.PeeringCreds.ClientSecret
	tenantId := state.tenantId

	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	resource, err := azureutil.ParseResourceID(obj.Status.RemoteId)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error parsing remote virtual network peering ID", composed.StopAndForget, ctx)
	}

	subscriptionId := resource.Subscription

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed to create azure loadRemoteVpcPeering client", composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx)
	}

	//NotFound
	peering, err := c.GetPeering(ctx, resource.ResourceGroup, resource.ResourceName, resource.SubResourceName)

	if azuremeta.IsNotFound(err) {

		if obj.Status.State == cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected {
			return composed.StopAndForget, nil
		}

		obj.Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcPeeringConnection,
				Message: "Remote VPC peering connection not found",
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading remote VPC Peering", ctx)
	}

	logger = logger.WithValues("remoteId", ptr.Deref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	return nil, ctx
}
