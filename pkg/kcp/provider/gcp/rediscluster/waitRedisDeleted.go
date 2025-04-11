package rediscluster

import (
	"context"
	"errors"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitRedisDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisCluster == nil {
		return nil, nil
	}

	if state.gcpRedisCluster.State != clusterpb.Cluster_DELETING {
		errorMsg := "error: unexpected gcp redis state"
		logger.Error(errors.New(errorMsg), errorMsg)
		redisCluster := st.Obj().(*cloudcontrolv1beta1.RedisCluster)
		redisCluster.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisCluster).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonUnknown,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	logger.Info("Cluster is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
