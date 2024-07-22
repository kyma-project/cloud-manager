package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func waitElastiCacheDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheCluster == nil {
		return nil, nil
	}

	cacheState := ptr.Deref(state.elastiCacheCluster.CacheClusterStatus, "")

	if cacheState != awsmeta.ElastiCache_DELETING && cacheState != awsmeta.ElastiCache_DELETED {
		errorMsg := fmt.Sprintf("Error: unexpected aws elasticache cluster state: %s", cacheState)
		redisInstance := st.Obj().(*v1beta1.RedisInstance)
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionTypeError,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	if cacheState == awsmeta.ElastiCache_DELETED {
		return nil, nil
	}

	logger.Info("Instance is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
