package vpcpeering

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func remoteClientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// remote client can't be created if remote network is not found
	if state.remoteNetwork == nil {
		return nil, nil
	}

	tenantId := state.remoteNetwork.Status.Network.Azure.TenantId
	if tenantId == "" {
		tenantId = state.Scope().Spec.Scope.Azure.TenantId
	}

	c, err := state.clientProvider(
		ctx,
		azureconfig.AzureConfig.PeeringCreds.ClientId,
		azureconfig.AzureConfig.PeeringCreds.ClientSecret,
		state.remoteNetworkId.Subscription,
		tenantId,
	)
	if err != nil {
		logger.
			WithValues(
				"azureRemoteTenant", tenantId,
				"azureRemoteSubscription", state.remoteNetworkId.Subscription,
			).
			Error(err, "Error creating remote Azure client for KCP VpcPeering")

		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Faile creating Azure client for tenant %s subscription %s", state.remoteNetworkId.Subscription, tenantId),
			}).
			ErrorLogMessage("Error patching KCP VpcPeering with error state after remote client creation failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())). // try again in 5mins
			Run(ctx, state)
	}

	state.remoteClient = c

	return nil, nil
}
