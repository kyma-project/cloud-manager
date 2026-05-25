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

func updateManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis == nil {
		return nil, ctx
	}

	// TLSVersion is hardcoded; this update is a no-op until it becomes user-configurable.
	desiredTLS := armredisenterprise.TLSVersionOne2
	currentTLS := armredisenterprise.TLSVersionOne2
	if state.managedRedis.Properties != nil && state.managedRedis.Properties.MinimumTLSVersion != nil {
		currentTLS = *state.managedRedis.Properties.MinimumTLSVersion
	}

	if desiredTLS == currentTLS {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Updating Azure Managed Redis TLSVersion", "name", obj.Name, "from", currentTLS, "to", desiredTLS)

	err := state.client.UpdateCluster(ctx, state.resourceGroupName, obj.Name, armredisenterprise.ClusterUpdate{
		Properties: &armredisenterprise.ClusterUpdateProperties{
			MinimumTLSVersion: &desiredTLS,
		},
	})
	if err != nil {
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to update Azure Managed Redis: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
