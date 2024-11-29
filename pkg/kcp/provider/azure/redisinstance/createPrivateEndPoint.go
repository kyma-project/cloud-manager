package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createPrivateEndPoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateEndPoint != nil && ptr.Deref(state.privateEndPoint.Properties.ProvisioningState, "") != armnetwork.ProvisioningStateFailed {
		return nil, nil
	}
	logger.Info("Creating Azure Private EndPoint")
	resourceGroupName := state.resourceGroupName
	privateEndpointInstanceName := state.ObjAsRedisInstance().Name
	subnetName := azurecommon.AzureCloudManagerResourceGroupName(state.Scope().Spec.Scope.Azure.VpcNetwork)
	redisInstanceResourceId := azureutil.NewRedisInstanceResourceId(
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		resourceGroupName,
		ptr.Deref(state.azureRedisInstance.Name, "")).String()
	subnetResourceId := azureutil.NewSubnetResourceId(
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		resourceGroupName,
		subnetName,
		subnetName).String()
	err := state.client.CreatePrivateEndPoint(
		ctx,
		resourceGroupName,
		privateEndpointInstanceName,
		armnetwork.PrivateEndpoint{
			Location: to.Ptr(state.Scope().Spec.Region),
			Properties: &armnetwork.PrivateEndpointProperties{
				Subnet: &armnetwork.Subnet{
					ID:   to.Ptr(subnetResourceId),
					Name: to.Ptr(state.IpRange().Spec.Network.Name),
				},
				PrivateLinkServiceConnections: []*armnetwork.PrivateLinkServiceConnection{
					{
						Name: to.Ptr(privateEndpointInstanceName),
						Properties: &armnetwork.PrivateLinkServiceConnectionProperties{
							PrivateLinkServiceID: to.Ptr(redisInstanceResourceId),
							GroupIDs:             []*string{ptr.To("redisCache")},
						},
					},
				},
			},
		},
	)
	if err != nil {
		logger.Error(err, "Error creating Azure PrivateEndpoint")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating Azure PrivateEndpoint: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure PrivateEndpoint creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
