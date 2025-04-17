package gcprediscluster

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	gcpRedisCluster := state.ObjAsGcpRedisCluster()

	if state.KcpGcpRedisCluster == nil {
		logger.Error(fmt.Errorf("kcpGcpRedisCluster not found"), "KcpGcpRedisCluster not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, nil
	}

	nodeType, err := redisTierToNodeTypeConverter(gcpRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to tier and memorySizeGb"
		logger.Error(err, errMsg, "redisTier", gcpRedisCluster.Spec.RedisTier)
		gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpGcpRedisCluster.Spec.NodeType = nodeType
	state.KcpGcpRedisCluster.Spec.ShardCount = gcpRedisCluster.Spec.ShardCount
	state.KcpGcpRedisCluster.Spec.ReplicasPerShard = gcpRedisCluster.Spec.ReplicasPerShard
	state.KcpGcpRedisCluster.Spec.RedisConfigs = gcpRedisCluster.Spec.RedisConfigs

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpGcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP GcpRedisCluster", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
}
