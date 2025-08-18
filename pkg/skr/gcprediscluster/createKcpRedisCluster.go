package gcprediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpGcpRedisCluster != nil {
		return nil, ctx
	}

	if state.SkrSubnet == nil {
		return composed.StopWithRequeue, nil
	}

	gcpRedisCluster := state.ObjAsGcpRedisCluster()

	nodeType, err := redisTierToNodeTypeConverter(gcpRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "failed to map redisTier to nodeType"
		logger.Error(err, errMsg, "redisTier", gcpRedisCluster.Spec.RedisTier)
		gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gcpRedisCluster.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      gcpRedisCluster.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: gcpRedisCluster.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.GcpRedisClusterSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: gcpRedisCluster.Namespace,
				Name:      gcpRedisCluster.Name,
			},
			Subnet: cloudcontrolv1beta1.GcpSubnetRef{
				Name: state.SkrSubnet.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			NodeType:         nodeType,
			ShardCount:       gcpRedisCluster.Spec.ShardCount,
			ReplicasPerShard: gcpRedisCluster.Spec.ReplicasPerShard,
			RedisConfigs:     gcpRedisCluster.Spec.RedisConfigs,
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpGcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP GcpRedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP GcpRedisCluster")

	gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(gcpRedisCluster).
		ErrorLogMessage("Error setting Creating state on GcpRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
