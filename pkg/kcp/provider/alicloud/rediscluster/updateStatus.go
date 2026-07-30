package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloud "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()

	// ConnectionDomain is only populated once the instance leaves Creating; guard
	// against writing a bare ":6379" endpoint if this runs before it is set.
	if state.instance.ConnectionDomain == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("ConnectionDomain is empty; instance endpoint not yet assigned"),
			"AliCloud r-kvstore cluster has no ConnectionDomain",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	discoveryEndpoint := fmt.Sprintf("%s:%d", state.instance.ConnectionDomain, state.instance.Port)
	changed := false
	if kcp.Status.DiscoveryEndpoint != discoveryEndpoint {
		kcp.Status.DiscoveryEndpoint = discoveryEndpoint
		changed = true
	}
	if kcp.Status.NodeType != state.instance.InstanceClass {
		kcp.Status.NodeType = state.instance.InstanceClass
		changed = true
	}
	if kcp.Status.ShardCount != state.instance.ShardCount {
		kcp.Status.ShardCount = state.instance.ShardCount
		changed = true
	}
	if kcp.Status.ReplicasPerShard != state.instance.ReadOnlyCount {
		kcp.Status.ReplicasPerShard = state.instance.ReadOnlyCount
		changed = true
	}

	// AuthString was written at CreateInstance time. If it is missing now the
	// password is unrecoverable - surface an error rather than silently proceeding.
	if kcp.Status.AuthString == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("AuthString is empty; password was never persisted or was lost"),
			"AliCloud r-kvstore cluster has no AuthString",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	if kcp.Status.CaCert == "" {
		cert, err := alicloud.FetchApsaraDBCACert(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error fetching ApsaraDB CA cert for AliCloud r-kvstore cluster",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
		}
		kcp.Status.CaCert = cert
		changed = true
	}

	hasReady := meta.FindStatusCondition(kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyState := kcp.Status.State == cloudcontrolv1beta1.StateReady
	// A lingering Updating condition must not coexist with Ready. If one is present
	// fall through to SetExclusiveConditions below, which clears it.
	hasUpdating := meta.FindStatusCondition(kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating) != nil

	if !changed && hasReady && hasReadyState && !hasUpdating {
		return composed.StopAndForget, ctx
	}
	kcp.Status.State = cloudcontrolv1beta1.StateReady

	return composed.UpdateStatus(kcp).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "AliCloud Redis cluster is ready",
		}).
		ErrorLogMessage("Error updating KCP RedisCluster (alicloud) status").
		SuccessLogMsg("KCP RedisCluster (alicloud) is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
