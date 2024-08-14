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
		ptr.Deref(state.elastiCacheCluster.NodeGroups[0].PrimaryEndpoint.Address, ""),
		ptr.Deref(state.elastiCacheCluster.NodeGroups[0].PrimaryEndpoint.Port, 0),
	)

	readEndpoint := fmt.Sprintf("%s:%d",
		ptr.Deref(state.elastiCacheCluster.NodeGroups[0].ReaderEndpoint.Address, ""),
		ptr.Deref(state.elastiCacheCluster.NodeGroups[0].ReaderEndpoint.Port, 0),
	)

	redisInstance.Status.PrimaryEndpoint = primaryEndpoint
	redisInstance.Status.ReadEndpoint = readEndpoint

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
