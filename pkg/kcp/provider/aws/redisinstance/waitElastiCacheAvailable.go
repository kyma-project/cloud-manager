package redisinstance

import (
	"context"
	"errors"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func waitElastiCacheAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		errorMsg := "Error: elasti cache cluster instance is not loaded"
		logger.Error(errors.New(errorMsg), errorMsg)
		redisInstance := st.Obj().(*v1beta1.RedisInstance)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonUnknown,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	cacheState := ptr.Deref(state.elastiCacheReplicationGroup.Status, "")
	if cacheState == awsmeta.ElastiCache_AVAILABLE {
		return nil, nil
	}

	logger.Info("Redis instance is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
