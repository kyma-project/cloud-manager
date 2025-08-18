package gcprediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	gcpRedisCluster := state.ObjAsGcpRedisCluster()

	kcpCondErr := meta.FindStatusCondition(state.KcpGcpRedisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpGcpRedisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	kcpCondUpdating := meta.FindStatusCondition(state.KcpGcpRedisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating)
	kcpHasUpdatingCondition := kcpCondUpdating != nil

	skrCondErr := meta.FindStatusCondition(gcpRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(gcpRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	skrHasUpdatingCondition := meta.FindStatusCondition(gcpRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeUpdating) != nil

	if kcpHasUpdatingCondition && skrCondErr == nil && !skrHasUpdatingCondition {
		gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateUpdating
		return composed.UpdateStatus(gcpRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeUpdating,
				Message: kcpCondUpdating.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisCluster status with updating conditions").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if kcpCondErr != nil && skrCondErr == nil {
		gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady, cloudresourcesv1beta1.ConditionTypeUpdating).
			ErrorLogMessage("Error: updating GcpRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR GcpRedisCluster status with Ready condition")
		gcpRedisCluster.Status.State = cloudresourcesv1beta1.StateReady
		return composed.UpdateStatus(gcpRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionTypeUpdating).
			ErrorLogMessage("Error updating SKR GcpRedisCluster status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, ctx
}
