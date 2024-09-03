package awsredisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
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

	state.KcpRedisInstance.Spec.Instance.Aws.Parameters = awsRedisInstance.Spec.Parameters
	state.KcpRedisInstance.Spec.Instance.Aws.CacheNodeType = awsRedisInstance.Spec.CacheNodeType
	state.KcpRedisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade = awsRedisInstance.Spec.AutoMinorVersionUpgrade

	err := state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
