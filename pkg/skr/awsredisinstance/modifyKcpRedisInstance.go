package awsredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	awsRedisInstance := state.ObjAsAwsRedisInstance()

	if state.KcpRedisInstance == nil {
		logger.Error(fmt.Errorf("kcpRedisInstance not found"), "KcpRedisInstance not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, nil
	}

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisInstance.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to cacheNodeType"
		logger.Error(err, errMsg, "redisTier", awsRedisInstance.Spec.RedisTier)
		awsRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(awsRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AwsRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AwsRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisInstance.Spec.Instance.Aws.Parameters = awsRedisInstance.Spec.Parameters
	state.KcpRedisInstance.Spec.Instance.Aws.CacheNodeType = cacheNodeType
	state.KcpRedisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade = awsRedisInstance.Spec.AutoMinorVersionUpgrade
	state.KcpRedisInstance.Spec.Instance.Aws.AuthEnabled = awsRedisInstance.Spec.AuthEnabled
	state.KcpRedisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow = awsRedisInstance.Spec.PreferredMaintenanceWindow
	state.KcpRedisInstance.Spec.Instance.Aws.EngineVersion = awsRedisInstance.Spec.EngineVersion

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
