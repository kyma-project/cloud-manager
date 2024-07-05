package gcpredisinstance

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

	gcpRedisInstance := state.ObjAsGcpRedisInstance()

	kcpCondErr := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(gcpRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(gcpRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	if kcpCondErr != nil && skrCondErr == nil {
		gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR GcpRedisInstance status with Ready condition")
		gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateReady
		return composed.UpdateStatus(gcpRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating SKR GcpRedisInstance status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
