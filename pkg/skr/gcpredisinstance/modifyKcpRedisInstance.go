package gcpredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	_, memorySizeGb, err := redisTierToTierAndMemorySizeConverter(gcpRedisInstance.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to tier and memorySizeGb"
		logger.Error(err, errMsg, "redisTier", gcpRedisInstance.Spec.RedisTier)
		gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisInstance.Spec.Instance.Gcp.MemorySizeGb = memorySizeGb
	state.KcpRedisInstance.Spec.Instance.Gcp.RedisConfigs = gcpRedisInstance.Spec.RedisConfigs
	state.KcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy = toGcpMaintenancePolicy(gcpRedisInstance.Spec.MaintenancePolicy)
	state.KcpRedisInstance.Spec.Instance.Gcp.AuthEnabled = gcpRedisInstance.Spec.AuthEnabled

	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
}
