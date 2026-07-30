package alicloudrediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisCluster != nil {
		return nil, ctx
	}

	alicloudRedisCluster := state.ObjAsAlicloudRedisCluster()

	instanceClass, err := redisTierToInstanceClass(alicloudRedisCluster.Spec.RedisTier, alicloudRedisCluster.Spec.ShardCount)
	if err != nil {
		errMsg := "failed to map redisTier to instanceClass"
		logger.Error(err, errMsg, "redisTier", alicloudRedisCluster.Spec.RedisTier)
		alicloudRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(alicloudRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AlicloudRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and stopped SKR AlicloudRedisCluster status with Error condition").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.KcpRedisCluster = &cloudcontrolv1beta1.RedisCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      alicloudRedisCluster.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      alicloudRedisCluster.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: alicloudRedisCluster.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisClusterSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: alicloudRedisCluster.Namespace,
				Name:      alicloudRedisCluster.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.RedisClusterInfo{
				Alicloud: &cloudcontrolv1beta1.RedisClusterAlicloud{
					InstanceClass:    instanceClass,
					EngineVersion:    alicloudRedisCluster.Spec.EngineVersion,
					ShardCount:       alicloudRedisCluster.Spec.ShardCount,
					ReplicasPerShard: alicloudRedisCluster.Spec.ReplicasPerShard,
					Parameters:       alicloudRedisCluster.Spec.Parameters,
				},
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisCluster")

	alicloudRedisCluster.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(alicloudRedisCluster).
		ErrorLogMessage("Error setting Creating state on AlicloudRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
