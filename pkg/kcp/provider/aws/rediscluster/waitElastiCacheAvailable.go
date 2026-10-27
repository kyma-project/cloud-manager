package rediscluster

import (
	"context"
	"errors"
	"fmt"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
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
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisCluster)
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
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisCluster)

		events, evErr := state.awsClient.DescribeElastiCacheReplicationGroupEvents(ctx, ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, ""))
		reason := ""
		if evErr != nil {
			logger.Error(evErr, "failed to fetch ElastiCache replication group events")
		} else {
			reason = latestElastiCacheEventReason(events)
		}

		msg := "Failed to provision RedisCluster"
		if reason != "" {
			msg += ": " + reason
		}

		logger.Error(
			fmt.Errorf("ElastiCache replication group is in an unrecoverable state"),
			"ElastiCache replication group provisioning failed",
			"reason", reason,
		)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: msg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("RedisCluster failed to provision and cannot recover in place").
			Run(ctx, st)
	}

	if state.IsReplicationGroupAvailable() {
		return nil, ctx
	}

	// Available-but-no-members-yet is a transient eventual-consistency window, not
	// a terminal state - requeue and let the members appear.
	logger.Info("Redis instance is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}

// latestElastiCacheEventReason returns the Message of the most-recently dated
// event; undated events rank below dated ones, and when none are dated the last
// element is used. Returns "" for empty input.
func latestElastiCacheEventReason(events []elasticachetypes.Event) string {
	if len(events) == 0 {
		return ""
	}
	latest := events[len(events)-1]
	for _, e := range events {
		if e.Date != nil && (latest.Date == nil || e.Date.After(*latest.Date)) {
			latest = e
		}
	}
	return ptr.Deref(latest.Message, "")
}
