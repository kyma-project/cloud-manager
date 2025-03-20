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

func initLocalPeeringClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var auxiliaryTenants []string

	if state.remoteNetwork != nil && state.remoteNetwork.Status.Network.Azure.TenantId != "" {
		auxiliaryTenants = append(auxiliaryTenants, state.remoteNetwork.Status.Network.Azure.TenantId)
	}

	// This client uses auxiliary tenant to authenticate in remote tenant and fails if SPN does not exist
	// in remote tenant. This client should only be used for local CreatePeering API call.
	client, err := state.clientProvider(ctx,
		azureconfig.AzureConfig.PeeringCreds.ClientId,
		azureconfig.AzureConfig.PeeringCreds.ClientSecret,
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		state.Scope().Spec.Scope.Azure.TenantId,
		auxiliaryTenants...)

	if err != nil {
		logger.Error(err, "Error creating local peering Azure client for KCP VpcPeering")

		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:   cloudcontrolv1beta1.ConditionTypeError,
				Status: metav1.ConditionTrue,
				Reason: cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed creating Azure peering client for tenant %s subscription %s",
					state.Scope().Spec.Scope.Azure.SubscriptionId,
					state.Scope().Spec.Scope.Azure.TenantId),
			}).
			ErrorLogMessage("Error patching KCP VpcPeering with error state after local peering client creation failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())). // try again in 5mins
			Run(ctx, state)
	}

	state.localPeeringClient = client

	return nil, nil
}
