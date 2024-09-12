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
	remote, err := azureutil.ParseResourceID(obj.Spec.VpcPeering.Azure.RemoteVnet)

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
	virtualNetworkId := azureutil.NewVirtualNetworkResourceId(
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		state.Scope().Spec.Scope.Azure.VpcNetwork, // ResourceGroup name is the same as VPC network name.
		state.Scope().Spec.Scope.Azure.VpcNetwork).String()

	err = c.CreatePeering(ctx,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		virtualNetworkId,
		true,
	)

	if err != nil {
		logger.Error(err, "Error creating remote VPC Peering")

		message := azuremeta.GetErrorMessage(err)

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: message,
		}

		if !composed.AnyConditionChanged(obj, condition) {
			if obj.Status.State == cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected {
				return composed.StopAndForget, ctx
			}
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		return composed.UpdateStatus(obj).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure remote VPC Peering created")

	return nil, nil
}
