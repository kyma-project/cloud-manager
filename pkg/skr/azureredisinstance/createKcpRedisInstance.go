package azureredisinstance

import (
	"context"
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

	azureRedisInstance := state.ObjAsAzureRedisInstance()

	redisSKUCapacity, _ := RedisTierToSKUCapacityConverter(azureRedisInstance.Spec.RedisTier)

	state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      azureRedisInstance.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      azureRedisInstance.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: azureRedisInstance.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: azureRedisInstance.Namespace,
				Name:      azureRedisInstance.Name,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Azure: &cloudcontrolv1beta1.RedisInstanceAzure{
					SKU:                cloudcontrolv1beta1.AzureRedisSKU{Capacity: redisSKUCapacity},
					RedisVersion:       azureRedisInstance.Spec.RedisVersion,
					ShardCount:         azureRedisInstance.Spec.ShardCount,
					ReplicasPerPrimary: azureRedisInstance.Spec.ReplicasPerPrimary,
					RedisConfiguration: cloudcontrolv1beta1.RedisInstanceAzureConfigs{
						MaxClients:                     azureRedisInstance.Spec.RedisConfiguration.MaxClients,
						MaxFragmentationMemoryReserved: azureRedisInstance.Spec.RedisConfiguration.MaxFragmentationMemoryReserved,
						MaxMemoryDelta:                 azureRedisInstance.Spec.RedisConfiguration.MaxMemoryDelta,
						MaxMemoryPolicy:                azureRedisInstance.Spec.RedisConfiguration.MaxMemoryPolicy,
						MaxMemoryReserved:              azureRedisInstance.Spec.RedisConfiguration.MaxMemoryReserved,
						NotifyKeyspaceEvents:           azureRedisInstance.Spec.RedisConfiguration.NotifyKeyspaceEvents,
					},
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisInstance")

	azureRedisInstance.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(azureRedisInstance).
		ErrorLogMessage("Error setting Creating state on AzureRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
