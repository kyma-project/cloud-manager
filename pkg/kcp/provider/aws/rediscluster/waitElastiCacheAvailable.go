package rediscluster

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	cacheState := ptr.Deref(state.elastiCacheReplicationGroup.Status, "")

	if state.IsReplicationGroupAvailable() {
		return nil, ctx
	}

	switch awsmeta.EvaluateProvisioningAction(cacheState, st.Obj().GetCreationTimestamp().Time, time.Now()) {
	case awsmeta.ProvisioningActionHeal:
		reason := createFailedEventReason(ctx, state)
		logger.Info("auto-healing create-failed ElastiCache replication group", "reason", reason, "autoHeal", true)
		if err := state.awsClient.DeleteElastiCacheReplicationGroup(ctx, ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, "")); err != nil {
			return awsmeta.LogErrorAndReturn(err, "ElastiCache auto-heal delete failed", ctx)
		}
		redisCluster := st.Obj().(*cloudcontrolv1beta1.RedisCluster)
		return composed.UpdateStatus(redisCluster).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonProcessing,
				Message: "ElastiCache provisioning failed; deleting and recreating within the provisioning deadline",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			SuccessLogMsg("auto-healing create-failed ElastiCache replication group").
			Run(ctx, st)
	case awsmeta.ProvisioningActionTimeout:
		reason := ""
		if cacheState == awsmeta.ElastiCache_CREATE_FAILED {
			reason = createFailedEventReason(ctx, state)
		}
		msg := provisioningTimeoutMessage("RedisCluster", reason)
		logger.Error(
			fmt.Errorf("ElastiCache replication group provisioning timed out"),
			"RedisCluster provisioning deadline exceeded",
			"cacheState", cacheState,
		)
		redisCluster := st.Obj().(*cloudcontrolv1beta1.RedisCluster)
		redisCluster.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisCluster).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: msg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("RedisCluster provisioning deadline exceeded").
			Run(ctx, st)
	}

	// ProvisioningActionWait — still provisioning within the deadline
	logger.Info("RedisCluster is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}

// createFailedEventReason fetches events and returns the most recent event reason, or "" on error.
func createFailedEventReason(ctx context.Context, state *State) string {
	logger := composed.LoggerFromCtx(ctx)
	events, evErr := state.awsClient.DescribeElastiCacheReplicationGroupEvents(ctx, ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, ""))
	if evErr != nil {
		logger.Error(evErr, "failed to fetch ElastiCache replication group events")
		return ""
	}
	return latestElastiCacheEventReason(events)
}

// provisioningTimeoutMessage builds the terminal timeout condition message.
func provisioningTimeoutMessage(kind, reason string) string {
	msg := fmt.Sprintf("Failed to provision %s: provisioning did not complete within the %s deadline; delete and recreate this resource to retry",
		kind, awsmeta.ElastiCacheProvisioningTimeout)
	if reason != "" {
		msg += " (cause: " + reason + ")"
	}
	return msg
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
