package alicloudredisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisInstance != nil {
		return nil, ctx
	}

	alicloudRedisInstance := state.ObjAsAlicloudRedisInstance()

	instanceClass, readOnlyCount, err := redisTierToInstanceClassAndReadOnlyCount(alicloudRedisInstance.Spec.RedisTier)
	if err != nil {
		errMsg := "failed to map redisTier to instanceClass"
		logger.Error(err, errMsg, "redisTier", alicloudRedisInstance.Spec.RedisTier)
		alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(alicloudRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AlicloudRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and stopped SKR AlicloudRedisInstance status with Error condition").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      alicloudRedisInstance.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      alicloudRedisInstance.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: alicloudRedisInstance.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: alicloudRedisInstance.Namespace,
				Name:      alicloudRedisInstance.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Alicloud: &cloudcontrolv1beta1.RedisInstanceAlicloud{
					InstanceClass: instanceClass,
					EngineVersion: alicloudRedisInstance.Spec.EngineVersion,
					ReadOnlyCount: readOnlyCount,
					Parameters:    alicloudRedisInstance.Spec.Parameters,
				},
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisInstance")

	alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(alicloudRedisInstance).
		ErrorLogMessage("Error setting Creating state on AlicloudRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
