package managedredis

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createVirtualNetworkLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.virtualNetworkLink != nil {
		return nil, ctx
	}

	subscriptionId := state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId
	vnetId := azureutil.NewVirtualNetworkResourceId(subscriptionId, state.gardenerNetworkName, state.gardenerNetworkName).String()

	composed.LoggerFromCtx(ctx).Info("Creating Virtual Network Link for Azure Managed Redis", "zone", state.PrivateDNSZoneName(), "link", virtualNetworkLinkName(state))

	err := state.client.CreateVirtualNetworkLink(ctx, state.resourceGroupName, state.PrivateDNSZoneName(), virtualNetworkLinkName(state), vnetId)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error creating Virtual Network Link for Azure Managed Redis")
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Virtual Network Link: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
