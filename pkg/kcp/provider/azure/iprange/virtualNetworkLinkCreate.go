package iprange

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func virtualNetworkLinkCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink != nil {
		return nil, ctx
	}

	logger.Info("Creating Azure KCP IpRange virtualNetworkLink")

	virtualNetworkLinkName := state.ObjAsIpRange().Name
	resourceGroupName := state.resourceGroupName
	kymaNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	privateDnsZoneName := azureutil.NewPrivateDnsZoneName()
	virtualNetworkLink := armprivatedns.VirtualNetworkLink{
		Location: ptr.To("global"),
		Properties: &armprivatedns.VirtualNetworkLinkProperties{
			VirtualNetwork: &armprivatedns.SubResource{
				ID: ptr.To(azureutil.NewVirtualNetworkResourceId(state.Scope().Spec.Scope.Azure.SubscriptionId,
					state.Scope().Spec.Scope.Azure.VpcNetwork, kymaNetworkName).String()),
			},
			RegistrationEnabled: ptr.To(false),
		},
	}
	err := state.azureClient.CreateVirtualNetworkLink(ctx, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName, virtualNetworkLink)

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Azure KCP IpRange too many requests on virtual network link create",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	if err != nil {
		logger.Error(err, "Error creating Azure KCP IpRange virtualNetworkLink")

		state.ObjAsIpRange().Status.State = cloudcontrol1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrol1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrol1beta1.ConditionTypeError,
				Message: "Error creating Azure virtualNetworkLink",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after failed creating virtualNetworkLink").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
