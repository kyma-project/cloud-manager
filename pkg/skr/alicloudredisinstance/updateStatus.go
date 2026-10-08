package alicloudredisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisInstance == nil {
		return nil, ctx
	}

	alicloudRedisInstance := state.ObjAsAlicloudRedisInstance()

	kcpCondErr := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	kcpCondUpdating := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating)
	kcpHasUpdatingCondition := kcpCondUpdating != nil

	skrCondErr := meta.FindStatusCondition(alicloudRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(alicloudRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	skrHasUpdatingCondition := meta.FindStatusCondition(alicloudRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeUpdating) != nil

	if kcpHasUpdatingCondition && skrCondErr == nil && !skrHasUpdatingCondition {
		alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateUpdating
		return composed.UpdateStatus(alicloudRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeUpdating,
				Message: kcpCondUpdating.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AlicloudRedisInstance status with updating conditions").
			SuccessErrorNil().
			Run(ctx, state)
	}

	if kcpCondErr != nil && skrCondErr == nil {
		alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(alicloudRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady, cloudresourcesv1beta1.ConditionTypeUpdating).
			ErrorLogMessage("Error: updating AlicloudRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated SKR AlicloudRedisInstance status with Error condition, requeuing").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR AlicloudRedisInstance status with Ready condition")
		alicloudRedisInstance.Status.State = cloudresourcesv1beta1.StateReady
		return composed.UpdateStatus(alicloudRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError, cloudresourcesv1beta1.ConditionTypeUpdating).
			ErrorLogMessage("Error updating SKR AlicloudRedisInstance status with ready condition").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
