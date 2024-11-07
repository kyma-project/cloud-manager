package awsredisinstance

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisInstance != nil {
		return nil, nil
	}

	awsRedisInstance := state.ObjAsAwsRedisInstance()

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisInstance.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to cacheNodeType"
		logger.Error(errors.New(errMsg), errMsg, "redisTier", awsRedisInstance.Spec.RedisTier)
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

	replicaCount := redisTierToReadReplicas(awsRedisInstance.Spec.RedisTier)

	state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsRedisInstance.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      awsRedisInstance.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: awsRedisInstance.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: awsRedisInstance.Namespace,
				Name:      awsRedisInstance.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Aws: &cloudcontrolv1beta1.RedisInstanceAws{
					CacheNodeType:              cacheNodeType,
					EngineVersion:              awsRedisInstance.Spec.EngineVersion,
					AutoMinorVersionUpgrade:    awsRedisInstance.Spec.AutoMinorVersionUpgrade,
					AuthEnabled:                awsRedisInstance.Spec.AuthEnabled,
					PreferredMaintenanceWindow: awsRedisInstance.Spec.PreferredMaintenanceWindow,
					Parameters:                 awsRedisInstance.Spec.Parameters,
					ReadReplicas:               replicaCount,
				},
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisInstance")

	awsRedisInstance.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(awsRedisInstance).
		ErrorLogMessage("Error setting Creating state on AwsRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
