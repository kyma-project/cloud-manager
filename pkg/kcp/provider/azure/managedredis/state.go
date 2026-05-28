package managedredis

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremanagedredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

const (
	PrivateEndpointGroupID = "redisEnterprise"
	PrivateDNSZone         = "privatelink.redis.azure.net"
	PrivateDNSZoneChina    = "privatelink.redis.cache.chinacloudapi.cn"
	RedisPort              = int32(10000)
	DefaultDatabaseName    = "default"
)

type State struct {
	kcpcommonaction.State

	client            azuremanagedredisclient.Client
	resourceGroupName string

	managedRedis         *armredisenterprise.Cluster
	managedRedisDatabase *armredisenterprise.Database
	privateEndpoint      *armnetwork.PrivateEndpoint
	privateDnsZoneGroup  *armnetwork.PrivateDNSZoneGroup
}

func (s *State) ObjAsAzureManagedRedis() *cloudcontrolv1beta1.AzureManagedRedis {
	return s.Obj().(*cloudcontrolv1beta1.AzureManagedRedis)
}

func (s *State) PrivateDNSZoneName() string {
	if azureconfig.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		return PrivateDNSZoneChina
	}
	return PrivateDNSZone
}

// initAzureClient finalizes State by creating the Azure client and setting the
// resource group name from the loaded VpcNetwork and Subscription.
func initAzureClient(clientProvider azureclient.ClientProvider[azuremanagedredisclient.Client]) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		if state.client != nil {
			return nil, ctx
		}

		obj := state.ObjAsAzureManagedRedis()
		gardenerNetworkName := ptr.Deref(state.VpcNetwork().Spec.VpcNetworkName, "")
		if gardenerNetworkName == "" {
			obj.Status.State = string(cloudcontrolv1beta1.StateError)
			return composed.UpdateStatus(obj).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonInvalidDependency,
					Message: fmt.Sprintf("VpcNetwork %s has no spec.vpcNetworkName", state.VpcNetwork().Name),
				}).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
				Run(ctx, st)
		}

		clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
		clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
		subscriptionId := state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId
		tenantId := state.Subscription().Status.SubscriptionInfo.Azure.TenantId

		c, err := clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating Azure ManagedRedis client", composed.StopWithRequeue, ctx)
		}

		state.client = c
		state.resourceGroupName = azurecommon.AzureCloudManagerResourceGroupName(gardenerNetworkName)
		return nil, ctx
	}
}
