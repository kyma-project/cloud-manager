package gcpnfsvolume

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	condErr := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	condReady := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	if condErr != nil {
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: condErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with Error condition due to KCP error").
			Run(ctx, state)
	}

	if condReady != nil {
		state.ObjAsGcpNfsVolume().Status.CapacityGb = state.KcpNfsInstance.Status.CapacityGb
		state.ObjAsGcpNfsVolume().Status.Hosts = state.KcpNfsInstance.Status.Hosts
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: condReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating GcpNfsVolume status with Ready condition").
			SuccessLogMsg("GcpNfsVolume status got updated with Ready condition "+strconv.Itoa(state.KcpNfsInstance.Spec.Instance.Gcp.CapacityGb)).
			Run(ctx, state)
	}

	return nil, nil
}
