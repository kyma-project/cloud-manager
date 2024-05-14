package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func createVpcPeeringRemote(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	resource, err := util.ParseResourceID(obj.Spec.VpcPeering.Azure.RemoteVnet)

	if err != nil {
		logger.Error(err, "Error parsing remoteVnet")
	}

	clientId := azureconfig.AzureConfig.ClientId
	clientSecret := azureconfig.AzureConfig.ClientSecret
	subscriptionId := resource.Subscription
	tenantId := azureconfig.AzureConfig.TenantId

	c, err := state.provider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return err, ctx
	}

	resourceGroupName := obj.Spec.VpcPeering.Azure.RemoteResourceGroup
	virtualNetworkPeeringName := fmt.Sprintf("%s-%s", obj.Spec.RemoteRef.Namespace, obj.Spec.RemoteRef.Name)

	remoteVnet := "/subscriptions/9c05f3c1-314b-4c4b-bfff-b5a0650177cb/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet"

	peering, err := c.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		virtualNetworkPeeringName,
		remoteVnet,
		obj.Spec.VpcPeering.Azure.AllowVnetAccess,
	)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Failed creating VpcPeerings %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	// TODO should we have different logger values for different providers like connectionId and ID
	logger = logger.WithValues("remotePeeringId", pointer.StringDeref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("Azure remote VPC Peering created")

	return nil, ctx
}
