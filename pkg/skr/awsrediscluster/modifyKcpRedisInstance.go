package awsrediscluster

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	awsRedisCluster := state.ObjAsAwsRedisCluster()

	if state.KcpRedisCluster == nil {
		logger.Error(fmt.Errorf("kcpRedisCluster not found"), "KcpRedisCluster not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, nil
	}

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to cacheNodeType"
		logger.Error(err, errMsg, "redisTier", awsRedisCluster.Spec.RedisTier)
		awsRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(awsRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AwsRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AwsRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisCluster.Spec.Instance.Aws.Parameters = awsRedisCluster.Spec.Parameters
	state.KcpRedisCluster.Spec.Instance.Aws.CacheNodeType = cacheNodeType
	state.KcpRedisCluster.Spec.Instance.Aws.AutoMinorVersionUpgrade = awsRedisCluster.Spec.AutoMinorVersionUpgrade
	state.KcpRedisCluster.Spec.Instance.Aws.AuthEnabled = awsRedisCluster.Spec.AuthEnabled
	state.KcpRedisCluster.Spec.Instance.Aws.PreferredMaintenanceWindow = awsRedisCluster.Spec.PreferredMaintenanceWindow
	state.KcpRedisCluster.Spec.Instance.Aws.EngineVersion = awsRedisCluster.Spec.EngineVersion
	state.KcpRedisCluster.Spec.Instance.Aws.ShardCount = awsRedisCluster.Spec.ShardCount
	state.KcpRedisCluster.Spec.Instance.Aws.ReplicasPerShard = awsRedisCluster.Spec.ReplicasPerShard

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
