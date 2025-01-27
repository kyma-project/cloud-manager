package redisinstance

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()
	hasChanged := false

	primaryEndpoint := fmt.Sprintf(
		"%s:%d",
		*state.azureRedisInstance.Properties.HostName,
		*state.azureRedisInstance.Properties.SSLPort,
	)
	if redisInstance.Status.PrimaryEndpoint != primaryEndpoint {
		redisInstance.Status.PrimaryEndpoint = primaryEndpoint
		hasChanged = true
	}
	resourceGroupName := state.resourceGroupName
	keys, err := state.client.GetRedisInstanceAccessKeys(ctx, resourceGroupName, state.ObjAsRedisInstance().Name)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error retrieving Azure RedisInstance access keys", composed.StopWithRequeue, ctx)
	}

	authString := ""
	if state.azureRedisInstance != nil {
		authString = pie.First(keys)
	}
	if redisInstance.Status.AuthString != authString {
		redisInstance.Status.AuthString = authString
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
