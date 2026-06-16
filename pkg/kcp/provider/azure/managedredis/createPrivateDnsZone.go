package managedredis

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createPrivateDnsZone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.privateDnsZone != nil {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Creating Private DNS Zone for Azure Managed Redis", "zone", state.PrivateDNSZoneName())

	err := state.client.CreatePrivateDnsZone(ctx, state.resourceGroupName, state.PrivateDNSZoneName(), nil)
	if err != nil {
		// 409 Conflict on the shared private DNS zone happens when multiple AMR
		// reconciles race on the same `privatelink.redis.azure.net` zone — Azure
		// serializes upserts and rejects concurrent requests with a queued-operation
		// 409. Treat it as transient: loadPrivateDnsZone will pick the zone up on
		// the next reconcile once the in-flight upsert finishes.
		if azuremeta.IsConflictError(err) {
			composed.LoggerFromCtx(ctx).Info("Private DNS Zone create conflicted with concurrent operation; will retry", "zone", state.PrivateDNSZoneName())
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}
		composed.LoggerFromCtx(ctx).Error(err, "Error creating Private DNS Zone for Azure Managed Redis")
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Private DNS Zone: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
