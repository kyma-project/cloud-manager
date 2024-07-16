package gcpvpcpeering

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

	if state.KcpVpcPeering == nil {
		// it's deleted
		return nil, nil
	}

	kcpCondErr := meta.FindStatusCondition(state.KcpVpcPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpVpcPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(state.ObjAsGcpVpcPeering().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(state.ObjAsGcpVpcPeering().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	if kcpCondErr != nil && skrCondErr == nil {
		return composed.UpdateStatus(state.ObjAsGcpVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating KCP GcpVpcPeering status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpVpcPeering status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR GcpVpcPeering status with Ready condition")

		return composed.UpdateStatus(state.ObjAsGcpVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating KCP GcpVpcPeering status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if state.KcpVpcPeering.Status.Conditions[0].LastTransitionTime.Time.After(state.ObjAsGcpVpcPeering().Status.Conditions[0].LastTransitionTime.Time) &&
		state.KcpVpcPeering.Status.Conditions[0].Message != state.ObjAsGcpVpcPeering().Status.Conditions[0].Message &&
		kcpCondReady != nil &&
		skrCondReady != nil {
		return composed.UpdateStatus(state.ObjAsGcpVpcPeering()).
			SetExclusiveConditions(state.KcpVpcPeering.Status.Conditions[0]).
			ErrorLogMessage(state.KcpVpcPeering.Status.Conditions[0].Message).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
