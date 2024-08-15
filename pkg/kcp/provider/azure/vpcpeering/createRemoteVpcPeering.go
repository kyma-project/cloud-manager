package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func createRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remotePeering != nil {
		return nil, nil
	}

	clientId := azureconfig.AzureConfig.PeeringCreds.ClientId
	clientSecret := azureconfig.AzureConfig.PeeringCreds.ClientSecret
	tenantId := state.tenantId

	// We are creating virtual network peering in remote subscription therefore we are decomposing remoteVnetID
	remote, err := util.ParseResourceID(obj.Spec.VpcPeering.Azure.RemoteVnet)

	if err != nil {
		logger.Error(err, "Error parsing remoteVnet")
		return err, ctx
	}

	subscriptionId := remote.Subscription

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return err, ctx
	}

	virtualNetworkName := remote.ResourceName
	resourceGroupName := obj.Spec.VpcPeering.Azure.RemoteResourceGroup
	virtualNetworkPeeringName := obj.Spec.VpcPeering.Azure.RemotePeeringName

	// Since we are creating virtual network peering connection from remote to shoot we need to build shootNetworkID
	virtualNetworkId := util.VirtualNetworkResourceId(
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		state.Scope().Spec.Scope.Azure.VpcNetwork, // ResourceGroup name is the same as VPC network name.
		state.Scope().Spec.Scope.Azure.VpcNetwork)

	peering, err := c.CreatePeering(ctx,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		virtualNetworkId,
		obj.Spec.VpcPeering.Azure.AllowVnetAccess,
	)

	if err != nil {
		logger.Error(err, "Error creating remote VPC Peering")

		message := azuremeta.GetErrorMessage(err)

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: message,
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	logger = logger.WithValues("remotePeeringId", ptr.Deref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("Azure remote VPC Peering created")

	obj.Status.RemoteId = ptr.Deref(peering.ID, "")

	return composed.UpdateStatus(obj).
		ErrorLogMessage("Error updating VpcPeering status with remote connection id").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
