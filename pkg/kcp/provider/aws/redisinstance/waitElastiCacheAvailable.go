package redisinstance

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

	if state.IsReplicationGroupAvailable() {
		return nil, ctx
	}

	// AVAILABLE but member clusters not yet visible — transient eventual-consistency;
	// not subject to provisioning deadline.
	if cacheState == awsmeta.ElastiCache_AVAILABLE {
		logger.Info("Redis instance is not ready yet, requeueing with delay")
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	switch awsmeta.EvaluateProvisioningAction(cacheState, st.Obj().GetCreationTimestamp().Time, time.Now()) {
	case awsmeta.ProvisioningActionHeal:
		reason := createFailedEventReason(ctx, state)
		msg := "Failed to provision RedisInstance"
		if reason != "" {
			msg += ": " + reason
		}
		logger.Error(
			fmt.Errorf("ElastiCache replication group is in an unrecoverable state"),
			"ElastiCache replication group provisioning failed",
			"reason", reason,
		)
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: msg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("RedisInstance failed to provision and cannot recover in place").
			Run(ctx, st)
	case awsmeta.ProvisioningActionTimeout:
		reason := ""
		if cacheState == awsmeta.ElastiCache_CREATE_FAILED {
			reason = createFailedEventReason(ctx, state)
		}
		msg := provisioningTimeoutMessage("RedisInstance", reason)
		logger.Error(
			fmt.Errorf("ElastiCache replication group provisioning timed out"),
			"RedisInstance provisioning deadline exceeded",
			"cacheState", cacheState,
		)
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: msg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("RedisInstance provisioning deadline exceeded").
			Run(ctx, st)
	}

	// ProvisioningActionWait — still provisioning within the deadline
	logger.Info("Redis instance is not ready yet, requeueing with delay")
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
