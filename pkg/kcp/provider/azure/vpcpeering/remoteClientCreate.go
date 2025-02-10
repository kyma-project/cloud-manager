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

	var auxiliaryTenants []string

	localTenantId := state.Scope().Spec.Scope.Azure.TenantId

	remoteTenantId := state.remoteNetwork.Status.Network.Azure.TenantId

	// if remote tenant is not specified default it to local tenant
	if remoteTenantId == "" {
		remoteTenantId = state.Scope().Spec.Scope.Azure.TenantId
	}

	// if not on the same tenant add local tenant as auxiliary tenant
	if remoteTenantId != localTenantId {
		auxiliaryTenants = append(auxiliaryTenants, localTenantId)
	}

	client, err := state.clientProvider(
		ctx,
		azureconfig.AzureConfig.PeeringCreds.ClientId,
		azureconfig.AzureConfig.PeeringCreds.ClientSecret,
		state.remoteNetworkId.Subscription,
		remoteTenantId,
		auxiliaryTenants...,
	)
	if err != nil {
		logger.Error(err, "Error creating remote Azure client for KCP VpcPeering")

		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed creating Azure client for tenant %s subscription %s", state.remoteNetworkId.Subscription, remoteTenantId),
			}).
			ErrorLogMessage("Error patching KCP VpcPeering with error state after remote client creation failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())). // try again in 5mins
			Run(ctx, state)
	}

	state.remoteClient = client

	return nil, nil
}
