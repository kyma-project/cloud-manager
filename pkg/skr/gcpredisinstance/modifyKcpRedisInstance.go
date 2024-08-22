package gcpredisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func modifyKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	gcpRedisInstance := state.ObjAsGcpRedisInstance()

	if state.KcpRedisInstance == nil {
		logger.Error(fmt.Errorf("kcpRedisInstance not found"), "KcpRedisInstance not found")
		return composed.StopWithRequeue, nil
	}

	shouldModifyKcp := state.ShouldModifyKcp()

	if !shouldModifyKcp {
		return nil, nil
	}

	state.KcpRedisInstance.Spec.Instance.Gcp.MemorySizeGb = gcpRedisInstance.Spec.MemorySizeGb
	state.KcpRedisInstance.Spec.Instance.Gcp.RedisConfigs = gcpRedisInstance.Spec.RedisConfigs
	state.KcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy = toGcpMaintenancePolicy(gcpRedisInstance.Spec.MaintenancePolicy)
	state.KcpRedisInstance.Spec.Instance.Gcp.AuthEnabled = gcpRedisInstance.Spec.AuthEnabled

	err := state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
}
