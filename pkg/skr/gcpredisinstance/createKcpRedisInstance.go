package gcpredisinstance

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

	gcpRedisInstance := state.ObjAsGcpRedisInstance()

	state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gcpRedisInstance.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      gcpRedisInstance.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: gcpRedisInstance.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: gcpRedisInstance.Namespace,
				Name:      gcpRedisInstance.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{
					Tier:                  gcpRedisInstance.Spec.Tier,
					MemorySizeGb:          gcpRedisInstance.Spec.MemorySizeGb,
					RedisVersion:          gcpRedisInstance.Spec.RedisVersion,
					AuthEnabled:           gcpRedisInstance.Spec.AuthEnabled,
					TransitEncryptionMode: gcpRedisInstance.Spec.TransitEncryptionMode,
					RedisConfigs:          gcpRedisInstance.Spec.RedisConfigs,
					MaintenancePolicy:     toGcpMaintenancePolicy(gcpRedisInstance.Spec.MaintenancePolicy),
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP RedisInstance")

	gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.PatchStatus(gcpRedisInstance).
		ErrorLogMessage("Error setting Creating state on GcpRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
