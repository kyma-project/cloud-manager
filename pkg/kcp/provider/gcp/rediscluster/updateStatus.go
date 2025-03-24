package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsRedisCluster()

	hasChanged := false

	if len(state.gcpRedisCluster.DiscoveryEndpoints) > 0 {
		discoveryEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisCluster.DiscoveryEndpoints[0].Address, state.gcpRedisCluster.DiscoveryEndpoints[0].Port)
		if redisCluster.Status.DiscoveryEndpoint != discoveryEndpoint {
			redisCluster.Status.DiscoveryEndpoint = discoveryEndpoint
			hasChanged = true
		}
	}

	hasReadyCondition := meta.FindStatusCondition(redisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := redisCluster.Status.State == cloudcontrolv1beta1.StateReady

	if !hasChanged && hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("RedisCluster status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	redisCluster.Status.State = cloudcontrolv1beta1.StateReady
	return composed.UpdateStatus(redisCluster).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Redis instance is ready",
		}).
		ErrorLogMessage("Error updating KCP RedisCluster status after setting Ready condition").
		SuccessLogMsg("KCP RedisCluster is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
