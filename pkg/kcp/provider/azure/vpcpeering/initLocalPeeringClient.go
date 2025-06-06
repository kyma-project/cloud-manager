package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"k8s.io/apimachinery/pkg/api/meta"
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

	if err == nil {
		state.localPeeringClient = client
		return nil, ctx
	}

	logger.Error(err, "Error creating local peering Azure client for KCP VpcPeering")

	changed := false

	if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateError) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
		changed = true
	}

	if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		changed = true
	}

	condition := metav1.Condition{
		Type:   cloudcontrolv1beta1.ConditionTypeError,
		Status: metav1.ConditionTrue,
		Reason: cloudcontrolv1beta1.ReasonCloudProviderError,
		Message: fmt.Sprintf("Failed creating Azure peering client for tenant %s subscription %s",
			state.Scope().Spec.Scope.Azure.SubscriptionId,
			state.Scope().Spec.Scope.Azure.TenantId),
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), condition) {
		changed = true
	}

	successError := composed.StopAndForget

	if !changed {
		return successError, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error patching KCP VpcPeering with error state after local peering client creation failed").
		SuccessError(successError). // try again in 5mins
		Run(ctx, state)

}
