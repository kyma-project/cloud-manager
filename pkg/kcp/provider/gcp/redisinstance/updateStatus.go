package redisinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	redisInstance := state.ObjAsRedisInstance()

	redisInstance.Status.PrimaryEndpoint = fmt.Sprintf("%s:%d", state.gcpRedisInstance.Host, state.gcpRedisInstance.Port)
	if state.gcpRedisInstance.ReadEndpoint != "" {
		redisInstance.Status.ReadEndpoint = fmt.Sprintf("%s:%d", state.gcpRedisInstance.ReadEndpoint, state.gcpRedisInstance.ReadEndpointPort)
	}
	if state.gcpRedisInstanceAuth != nil {
		redisInstance.Status.AuthString = state.gcpRedisInstanceAuth.AuthString
	}

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
