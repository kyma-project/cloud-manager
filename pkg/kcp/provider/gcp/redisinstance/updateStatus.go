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

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	redisInstance := state.ObjAsRedisInstance()

	primaryEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisInstance.Host, state.gcpRedisInstance.Port)
	if redisInstance.Status.PrimaryEndpoint != primaryEndpoint {
		redisInstance.Status.PrimaryEndpoint = primaryEndpoint
	}

	readEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisInstance.ReadEndpoint, state.gcpRedisInstance.ReadEndpointPort)
	if state.gcpRedisInstance.ReadEndpoint == "" {
		readEndpoint = ""
	}
	if redisInstance.Status.ReadEndpoint != readEndpoint {
		redisInstance.Status.ReadEndpoint = readEndpoint

	}

	authString := ""
	if state.gcpRedisInstanceAuth != nil {
		authString = state.gcpRedisInstanceAuth.AuthString
	}
	if redisInstance.Status.AuthString != authString {
		redisInstance.Status.AuthString = authString
	}

	if len(state.gcpRedisInstance.ServerCaCerts) > 0 && redisInstance.Status.CaCert != state.gcpRedisInstance.ServerCaCerts[0].Cert {
		redisInstance.Status.CaCert = state.gcpRedisInstance.ServerCaCerts[0].Cert
	}

	hasReadyCondition := meta.FindStatusCondition(redisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	if hasReadyCondition {
		composed.LoggerFromCtx(ctx).Info("Ready condition already present, StopAndForget-ing")
		return composed.StopAndForget, nil
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
