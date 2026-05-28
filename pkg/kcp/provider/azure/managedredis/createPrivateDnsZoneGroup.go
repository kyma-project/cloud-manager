package managedredis

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createPrivateDnsZoneGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.privateDnsZoneGroup != nil {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Creating Private DNS Zone Group", "name", obj.Name)

	dnsZoneId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/privateDnsZones/%s",
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId, state.resourceGroupName, state.PrivateDNSZoneName())
	configName := "default"
	dzgParams := armnetwork.PrivateDNSZoneGroup{
		Properties: &armnetwork.PrivateDNSZoneGroupPropertiesFormat{
			PrivateDNSZoneConfigs: []*armnetwork.PrivateDNSZoneConfig{
				{
					Name: &configName,
					Properties: &armnetwork.PrivateDNSZonePropertiesFormat{
						PrivateDNSZoneID: &dnsZoneId,
					},
				},
			},
		},
	}

	err := state.client.CreatePrivateDnsZoneGroup(ctx, state.resourceGroupName, obj.Name+"-pe", obj.Name+"-dzg", dzgParams)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error creating Private DNS Zone Group for Azure Managed Redis")
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Private DNS Zone Group: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
