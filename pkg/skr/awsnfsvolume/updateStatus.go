package awsnfsvolume

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

	kcpCondErr := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(state.ObjAsAwsNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(state.ObjAsAwsNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	if kcpCondErr != nil && skrCondErr == nil {
		logger.Info("Updating SKR AwsNfsVolume status with Error condition")
		return composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating KCP AwsNfsVolume status with not ready condition due to KCP error").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR AwsNfsVolume status with Ready condition")
		if len(state.KcpNfsInstance.Status.Hosts) > 0 {
			state.ObjAsAwsNfsVolume().Status.Server = state.KcpNfsInstance.Status.Hosts[0]
		}
		return composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating KCP AwsNfsVolume status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
