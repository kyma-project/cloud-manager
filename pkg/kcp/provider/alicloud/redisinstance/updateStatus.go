package redisinstance

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

// updateStatus writes the observed connection details onto the KCP object
// and marks the Ready condition true.
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisInstance()

	primaryEndpoint := fmt.Sprintf("%s:%d", state.instance.ConnectionDomain, state.instance.Port)
	changed := false
	if kcp.Status.PrimaryEndpoint != primaryEndpoint {
		kcp.Status.PrimaryEndpoint = primaryEndpoint
		changed = true
	}
	if kcp.Status.NodeType != state.instance.InstanceClass {
		kcp.Status.NodeType = state.instance.InstanceClass
		changed = true
	}
	// AliCloud Capacity is in MB; all supported instance classes are ≥1 GB so
	// integer division is exact.
	memGB := int32(state.instance.Capacity / 1024)
	if memGB != kcp.Status.MemorySizeGb {
		kcp.Status.MemorySizeGb = memGB
		changed = true
	}
	if kcp.Status.ReplicaCount != state.instance.ReadOnlyCount {
		kcp.Status.ReplicaCount = state.instance.ReadOnlyCount
		changed = true
	}

	// AuthString was written at CreateInstance time. If it is missing now the
	// password is unrecoverable - surface an error rather than silently proceeding.
	if kcp.Status.AuthString == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("AuthString is empty; password was never persisted or was lost"),
			"AliCloud r-kvstore instance has no AuthString",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	if kcp.Status.CaCert == "" {
		cert, err := alicloud.FetchApsaraDBCACert(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error fetching ApsaraDB CA cert for AliCloud r-kvstore instance",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
		}
		kcp.Status.CaCert = cert
		changed = true
	}

	hasReady := meta.FindStatusCondition(kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyState := kcp.Status.State == cloudcontrolv1beta1.StateReady

	if !changed && hasReady && hasReadyState {
		return composed.StopAndForget, ctx
	}
	kcp.Status.State = cloudcontrolv1beta1.StateReady

	return composed.UpdateStatus(kcp).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "AliCloud Redis instance is ready",
		}).
		ErrorLogMessage("Error updating KCP RedisInstance (alicloud) status").
		SuccessLogMsg("KCP RedisInstance (alicloud) is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
