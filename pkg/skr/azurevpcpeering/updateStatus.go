package azurevpcpeering

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
	obj := state.ObjAsAzureVpcPeering()

	if state.KcpVpcPeering == nil {
		// it's deleted
		return nil, nil
	}

	kcpCondErr := meta.FindStatusCondition(state.KcpVpcPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpVpcPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(state.ObjAsAzureVpcPeering().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(state.ObjAsAzureVpcPeering().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	if kcpCondErr != nil && (skrCondErr == nil || skrCondErr.Message != kcpCondErr.Message) {
		obj.Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating SKR AzureVpcPeering status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AzureVpcPeering status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR AzureVpcPeering status with Ready condition")

		obj.Status.State = state.KcpVpcPeering.Status.State
		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating SKR AzureVpcPeering status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
