package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisCluster()
	hasChanged := false

	discoveryEndpoint := fmt.Sprintf("%s:%d",
		ptr.Deref(state.elastiCacheReplicationGroup.ConfigurationEndpoint.Address, ""),
		ptr.Deref(state.elastiCacheReplicationGroup.ConfigurationEndpoint.Port, 0),
	)

	if redisInstance.Status.DiscoveryEndpoint != discoveryEndpoint {
		redisInstance.Status.DiscoveryEndpoint = discoveryEndpoint
		hasChanged = true
	}

	authString := ""
	if state.authTokenValue != nil {
		authString = ptr.Deref(state.authTokenValue.SecretString, "")
	}
	if redisInstance.Status.AuthString != authString {
		redisInstance.Status.AuthString = authString
		hasChanged = true
	}

	hasReadyCondition := meta.FindStatusCondition(redisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := redisInstance.Status.State == cloudcontrolv1beta1.StateReady
	if !hasChanged && hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("RedisCluster status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	redisInstance.Status.State = cloudcontrolv1beta1.StateReady
	return composed.UpdateStatus(redisInstance).
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
