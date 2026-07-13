package azuremanagedredis

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// createKcpAzureManagedRedis materialises the SKR AzureManagedRedis as a
// matching KCP AzureManagedRedis. The Kyma RedisTier letter is expanded via
// TierToSpec to fill SKU / HighAvailability / ClusteringPolicy on the KCP spec.
func createKcpAzureManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpAzureManagedRedis != nil {
		return nil, ctx
	}

	amr := state.ObjAsAzureManagedRedis()

	tierSpec, err := TierToSpec(amr.Spec.RedisTier)
	if err != nil {
		errMsg := "Failed to map AzureManagedRedis tier to AMR cluster spec"
		logger.Error(err, errMsg, "redisTier", amr.Spec.RedisTier)
		amr.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(amr).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			ErrorLogMessage("Failed to persist Error condition for unknown tier on AzureManagedRedis").
			SuccessLogMsg("Updated and forgot SKR AzureManagedRedis status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.KcpAzureManagedRedis = &cloudcontrolv1beta1.AzureManagedRedis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      amr.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      amr.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: amr.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.AzureManagedRedisSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: amr.Namespace,
				Name:      amr.Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			VpcNetwork: corev1.LocalObjectReference{
				Name: state.KymaRef.Name,
			},
			SKU:              string(tierSpec.SKU),
			HighAvailability: tierSpec.HighAvailability,
			ClusteringPolicy: string(tierSpec.ClusteringPolicy),
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, state.KcpAzureManagedRedis)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP AzureManagedRedis", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP AzureManagedRedis")

	amr.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(amr).
		ErrorLogMessage("Error setting Creating state on AzureManagedRedis").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
