package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisCluster := state.ObjAsGcpRedisCluster()

	readyCond := meta.FindStatusCondition(*redisCluster.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		return nil, nil
	}

	logger.Info("Removing Ready condition")

	meta.RemoveStatusCondition(redisCluster.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	redisCluster.Status.State = cloudcontrolv1beta1.StateDeleting
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating GcpRedisCluster status after removing Ready condition", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
