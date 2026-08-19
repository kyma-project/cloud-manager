package redisinstance

import (
	"context"
	"errors"
	"fmt"

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
		errorMsg := "error: elasti cache cluster instance is not loaded"
		logger.Error(errors.New(errorMsg), errorMsg)
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
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

	cacheState := ptr.Deref(state.elastiCacheReplicationGroup.Status, "")

	// create-failed is terminal - the replication group won't recover on its own.
	if cacheState == awsmeta.ElastiCache_CREATE_FAILED {
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
		logger.Error(
			fmt.Errorf("ElastiCache replication group is in an unrecoverable state: %q", cacheState),
			"ElastiCache replication group provisioning failed",
		)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: "Failed to provision RedisInstance",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("RedisInstance failed to provision and cannot recover in place").
			Run(ctx, st)
	}

	if cacheState == awsmeta.ElastiCache_AVAILABLE {
		// Available with no member clusters is an unrecoverable inconsistency
		// (an available replication group always has nodes) - surface an error
		// rather than proceed to operate on nothing.
		if len(state.memberClusters) == 0 {
			redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
			logger.Error(
				fmt.Errorf("replication group is %q but has no member clusters", cacheState),
				"ElastiCache replication group has no member clusters",
			)
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(redisInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: "Failed to provision RedisInstance",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg("RedisInstance is available but has no member clusters").
				Run(ctx, st)
		}
		return nil, ctx
	}

	logger.Info("Redis instance is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
