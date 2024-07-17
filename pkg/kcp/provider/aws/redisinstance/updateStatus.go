package redisinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	primaryEndpoint := fmt.Sprintf("%s:%d",
		ptr.Deref(state.elastiCacheCluster.CacheNodes[0].Endpoint.Address, ""),
		ptr.Deref(state.elastiCacheCluster.CacheNodes[0].Endpoint.Port, 0),
	)

	redisInstance.Status.PrimaryEndpoint = primaryEndpoint

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
