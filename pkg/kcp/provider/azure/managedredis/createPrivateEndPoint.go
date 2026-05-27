package managedredis

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// createPrivateEndPoint creates the Private Endpoint that connects the Azure Managed Redis
// cluster to the Kyma-managed VNet.
//
// Note on IpRange: AzureManagedRedis.Spec.IpRange is required and validated by
// kcpcommonaction.ipRangeLoad, but the subnet ID is derived directly from the VpcNetwork's
// gardener network name (the Kyma-managed resource group). This matches the pattern used by
// the existing pkg/kcp/provider/azure/redisinstance and rediscluster providers, where
// IpRange.Spec.Network is not consumed at the Azure API level. The IpRange field exists for
// cross-resource consistency (so users can model the network topology declaratively) and to
// surface missing-IpRange errors at admission time.
//
// Note on the recreate predicate: an existing Private Endpoint is left alone unless its
// provisioning state is Failed. Other in-flight states (Updating, Deleting) deliberately
// fall through to the no-op early-return; the waitPrivateEndPointAvailable /
// waitPrivateEndPointDeleted actions are responsible for polling those transitions.
func createPrivateEndPoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.privateEndpoint != nil &&
		state.privateEndpoint.Properties != nil &&
		ptr.Deref(state.privateEndpoint.Properties.ProvisioningState, "") != armnetwork.ProvisioningStateFailed {
		return nil, ctx
	}

	if state.managedRedis == nil || state.managedRedis.ID == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	composed.LoggerFromCtx(ctx).Info("Creating Private Endpoint for Azure Managed Redis", "name", obj.Name)

	subnetName := azurecommon.AzureCloudManagerResourceGroupName(ptr.Deref(state.VpcNetwork().Spec.VpcNetworkName, ""))
	subnetId := azureutil.NewSubnetResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
		state.resourceGroupName,
		subnetName,
		subnetName,
	).String()

	region := state.VpcNetwork().Spec.Region
	peConnName := obj.Name + "-pe-conn"
	groupID := PrivateEndpointGroupID
	peParams := armnetwork.PrivateEndpoint{
		Location: &region,
		Properties: &armnetwork.PrivateEndpointProperties{
			Subnet: &armnetwork.Subnet{
				ID: &subnetId,
			},
			PrivateLinkServiceConnections: []*armnetwork.PrivateLinkServiceConnection{
				{
					Name: &peConnName,
					Properties: &armnetwork.PrivateLinkServiceConnectionProperties{
						PrivateLinkServiceID: state.managedRedis.ID,
						GroupIDs:             []*string{&groupID},
					},
				},
			},
		},
	}

	err := state.client.CreatePrivateEndPoint(ctx, state.resourceGroupName, obj.Name+"-pe", peParams)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error creating Private Endpoint for Azure Managed Redis")
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Private Endpoint: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
