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

func initLocalClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// This client does not specify auxiliary tenants at initialization time, and therefore it is resilient to errors
	// when SPN does not exist in remote tenant.
	client, err := state.clientProvider(ctx,
		azureconfig.AzureConfig.PeeringCreds.ClientId,
		azureconfig.AzureConfig.PeeringCreds.ClientSecret,
		state.Scope().Spec.Scope.Azure.SubscriptionId,
		state.Scope().Spec.Scope.Azure.TenantId)

	if err == nil {
		state.localClient = client
		return nil, ctx
	}

	logger.Error(err, "Error creating local Azure client for KCP VpcPeering")

	state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)

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
		Message: fmt.Sprintf("Failed creating Azure client for tenant %s subscription %s",
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
		ErrorLogMessage("Error patching KCP VpcPeering with error state after local client creation failed").
		SuccessError(successError).
		Run(ctx, state)

}
