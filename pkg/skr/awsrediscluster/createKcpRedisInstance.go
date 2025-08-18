package awsrediscluster

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisCluster != nil {
		return nil, ctx
	}

	awsRedisCluster := state.ObjAsAwsRedisCluster()

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to cacheNodeType"
		logger.Error(errors.New(errMsg), errMsg, "redisTier", awsRedisCluster.Spec.RedisTier)
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

	state.KcpRedisCluster = &cloudcontrolv1beta1.RedisCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsRedisCluster.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      awsRedisCluster.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: awsRedisCluster.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisClusterSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: awsRedisCluster.Namespace,
				Name:      awsRedisCluster.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.RedisClusterInfo{
				Aws: &cloudcontrolv1beta1.RedisClusterAws{
					CacheNodeType:              cacheNodeType,
					EngineVersion:              awsRedisCluster.Spec.EngineVersion,
					AutoMinorVersionUpgrade:    awsRedisCluster.Spec.AutoMinorVersionUpgrade,
					AuthEnabled:                awsRedisCluster.Spec.AuthEnabled,
					PreferredMaintenanceWindow: awsRedisCluster.Spec.PreferredMaintenanceWindow,
					Parameters:                 awsRedisCluster.Spec.Parameters,
					ShardCount:                 awsRedisCluster.Spec.ShardCount,
					ReplicasPerShard:           awsRedisCluster.Spec.ReplicasPerShard,
				},
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisCluster")

	awsRedisCluster.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(awsRedisCluster).
		ErrorLogMessage("Error setting Creating state on AwsRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
