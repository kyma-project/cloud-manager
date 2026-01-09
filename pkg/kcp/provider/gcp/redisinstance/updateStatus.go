package redisinstance

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

	redisInstance := state.ObjAsRedisInstance()

	hasChanged := false

	primaryEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisInstance.Host, state.gcpRedisInstance.Port)
	if redisInstance.Status.PrimaryEndpoint != primaryEndpoint {
		redisInstance.Status.PrimaryEndpoint = primaryEndpoint
		hasChanged = true
	}

	readEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisInstance.ReadEndpoint, state.gcpRedisInstance.ReadEndpointPort)
	if state.gcpRedisInstance.ReadEndpoint == "" {
		readEndpoint = ""
	}
	if redisInstance.Status.ReadEndpoint != readEndpoint {
		redisInstance.Status.ReadEndpoint = readEndpoint
		hasChanged = true
	}

	authString := ""
	if state.gcpRedisInstanceAuth != nil {
		authString = state.gcpRedisInstanceAuth.AuthString
	}
	if redisInstance.Status.AuthString != authString {
		redisInstance.Status.AuthString = authString
		hasChanged = true
	}

	if len(state.gcpRedisInstance.ServerCaCerts) > 0 && redisInstance.Status.CaCert != state.gcpRedisInstance.ServerCaCerts[0].Cert {
		redisInstance.Status.CaCert = state.gcpRedisInstance.ServerCaCerts[0].Cert
		hasChanged = true
	}

	memorySizeGb := state.gcpRedisInstance.MemorySizeGb
	if redisInstance.Status.MemorySizeGb != memorySizeGb {
		redisInstance.Status.MemorySizeGb = memorySizeGb
		hasChanged = true
	}

	replicaCount := state.gcpRedisInstance.ReplicaCount
	if redisInstance.Status.ReplicaCount != replicaCount {
		redisInstance.Status.ReplicaCount = replicaCount
		hasChanged = true
	}

	hasReadyCondition := meta.FindStatusCondition(redisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := redisInstance.Status.State == cloudcontrolv1beta1.StateReady

	if !hasChanged && hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("RedisInstance status fields are already up-to-date, StopAndForget-ing")
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
		ErrorLogMessage("Error updating KCP RedisInstance status after setting Ready condition").
		SuccessLogMsg("KCP RedisInstance is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
