package managedredis

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitManagedRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	if state.managedRedis.Properties == nil || state.managedRedis.Properties.ProvisioningState == nil {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	provisioningState := *state.managedRedis.Properties.ProvisioningState

	resourceStateStr := ""
	if state.managedRedis.Properties.ResourceState != nil {
		resourceStateStr = string(*state.managedRedis.Properties.ResourceState)
	}

	// Failed and Canceled are terminal per the ARM async-operation contract.
	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/async-operations#provisioningstate-values
	if provisioningState == armredisenterprise.ProvisioningStateFailed ||
		provisioningState == armredisenterprise.ProvisioningStateCanceled {
		composed.LoggerFromCtx(ctx).
			WithValues(
				"provisioningState", string(provisioningState),
				"resourceState", resourceStateStr,
			).
			Info("Azure Managed Redis cluster reached terminal failure state")

		conditionMsg := fmt.Sprintf("Azure Managed Redis cluster provisioning %s (resourceState=%s)", provisioningState, resourceStateStr)

		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: conditionMsg,
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	if provisioningState != armredisenterprise.ProvisioningStateSucceeded {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, ctx
}
