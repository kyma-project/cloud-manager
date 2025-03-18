package rediscluster

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createPrivateDnsZoneGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateDnsZoneGroup != nil {
		return nil, nil
	}
	if state.privateEndPoint == nil {
		logger.Info("Can not create Azure Private DnsZone Group, Private EndPoint is not loaded, reque")
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}
	logger.Info("Creating Azure Private DnsZone")
	resourceGroupName := state.resourceGroupName
	privateDnsZoneGroupName := state.ObjAsRedisCluster().Name
	privateEndPointName := ptr.Deref(state.privateEndPoint.Name, "")
	privateDnsZoneClusterName := azureutil.NewPrivateDnsZoneName()
	subscriptionId := state.Scope().Spec.Scope.Azure.SubscriptionId
	privateDNSZoneGroup := armnetwork.PrivateDNSZoneGroup{
		Properties: &armnetwork.PrivateDNSZoneGroupPropertiesFormat{
			PrivateDNSZoneConfigs: []*armnetwork.PrivateDNSZoneConfig{
				{
					Name: ptr.To(privateDnsZoneClusterName),
					Properties: &armnetwork.PrivateDNSZonePropertiesFormat{
						PrivateDNSZoneID: ptr.To(azureutil.NewPrivateDnsZoneGroupResourceId(subscriptionId, resourceGroupName, privateDnsZoneClusterName).String()),
					},
				},
			},
		},
	}
	err := state.client.CreatePrivateDnsZoneGroup(ctx, resourceGroupName, privateEndPointName, privateDnsZoneGroupName, privateDNSZoneGroup)
	if err != nil {
		logger.Error(err, "Error creating Azure PrivateDnsZone")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating Azure PrivateDnsZone: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure PrivateDnsZone creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
