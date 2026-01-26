package azurerediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisCluster != nil {
		return nil, ctx
	}

	azureRedisCluster := state.ObjAsAzureRedisCluster()

	redisSKUCapacity, err := RedisTierToSKUCapacityConverter(azureRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "Failed to map redisTier to SKU Capacity"
		logger.Error(err, errMsg, "redisTier", azureRedisCluster.Spec.RedisTier)
		azureRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(azureRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AzureRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR azureRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpRedisCluster = &cloudcontrolv1beta1.RedisCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      azureRedisCluster.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      azureRedisCluster.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: azureRedisCluster.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisClusterSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: azureRedisCluster.Namespace,
				Name:      azureRedisCluster.Name,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Instance: cloudcontrolv1beta1.RedisClusterInfo{
				Azure: &cloudcontrolv1beta1.RedisClusterAzure{
					SKU:                cloudcontrolv1beta1.AzureRedisClusterSKU{Capacity: redisSKUCapacity},
					RedisVersion:       azureRedisCluster.Spec.RedisVersion,
					ShardCount:         int(azureRedisCluster.Spec.ShardCount),
					ReplicasPerPrimary: int(azureRedisCluster.Spec.ReplicasPerPrimary),
					RedisConfiguration: cloudcontrolv1beta1.RedisInstanceAzureConfigs{
						MaxClients:                     azureRedisCluster.Spec.RedisConfiguration.MaxClients,
						MaxFragmentationMemoryReserved: azureRedisCluster.Spec.RedisConfiguration.MaxFragmentationMemoryReserved,
						MaxMemoryDelta:                 azureRedisCluster.Spec.RedisConfiguration.MaxMemoryDelta,
						MaxMemoryPolicy:                azureRedisCluster.Spec.RedisConfiguration.MaxMemoryPolicy,
						MaxMemoryReserved:              azureRedisCluster.Spec.RedisConfiguration.MaxMemoryReserved,
						NotifyKeyspaceEvents:           azureRedisCluster.Spec.RedisConfiguration.NotifyKeyspaceEvents,
					},
				},
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisCluster")

	azureRedisCluster.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(azureRedisCluster).
		ErrorLogMessage("Error setting Creating state on AzureRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
