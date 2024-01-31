package iprange

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

	logger.Info("Updating IpRange status")

	condErr := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	condReady := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	if condErr != nil {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: condErr.Message,
			}).
			ErrorLogMessage("Error updating IpRange status with not ready condition due to KCP error").
			Run(ctx, state)
	}

	if condReady != nil {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: condReady.Message,
			}).
			ErrorLogMessage("Error updating IpRange status with ready condition").
			Run(ctx, state)
	}

	return nil, nil
}
