package dnsresolver

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func initRemoteClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var auxiliaryTenants []string

	localTenantId := state.Scope().Spec.Scope.Azure.TenantId

	remoteTenantId := state.ObjAsAzureVNetLink().Spec.RemoteTenant

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
		state.rulesetId.Subscription,
		remoteTenantId,
		auxiliaryTenants...,
	)

	if err == nil {
		state.remoteClient = client
		return nil, ctx
	}

	logger.Error(err, "Error creating remote Azure client for KCP AzureVNetLink")

	state.ObjAsAzureVNetLink().Status.State = string(cloudcontrolv1beta1.StateError)

	return composed.PatchStatus(state.ObjAsAzureVNetLink()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: fmt.Sprintf("Failed creating Azure client for tenant %s subscription %s", state.rulesetId.Subscription, remoteTenantId),
		}).
		ErrorLogMessage("Error patching KCP AzureVNetLink with error state after remote client creation failed").
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())). // try again in 5mins
		Run(ctx, state)

}
