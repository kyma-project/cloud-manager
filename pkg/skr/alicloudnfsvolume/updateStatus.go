package alicloudnfsvolume

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

	if state.KcpNfsInstance == nil {
		// it's deleted
		return nil, nil
	}

	kcpCondErr := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(state.ObjAsAlicloudNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(state.ObjAsAlicloudNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	capacityChanged := !state.ObjAsAlicloudNfsVolume().Status.Capacity.Equal(state.KcpNfsInstance.Status.Capacity)
	state.ObjAsAlicloudNfsVolume().Status.Capacity = state.KcpNfsInstance.Status.Capacity

	if kcpCondErr != nil && skrCondErr == nil {
		state.ObjAsAlicloudNfsVolume().Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(state.ObjAsAlicloudNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating KCP AlicloudNfsVolume status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AlicloudNfsVolume status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR AlicloudNfsVolume status with Ready condition")
		// AliCloud KCP NfsInstance exposes the mount target via the singular Status.Host.
		// Fall back to the deprecated Status.Hosts list for safety.
		if len(state.KcpNfsInstance.Status.Host) > 0 {
			state.ObjAsAlicloudNfsVolume().Status.Server = state.KcpNfsInstance.Status.Host
		} else if len(state.KcpNfsInstance.Status.Hosts) > 0 {
			state.ObjAsAlicloudNfsVolume().Status.Server = state.KcpNfsInstance.Status.Hosts[0]
		}
		state.ObjAsAlicloudNfsVolume().Status.State = cloudresourcesv1beta1.StateReady
		return composed.UpdateStatus(state.ObjAsAlicloudNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
			ErrorLogMessage("Error updating KCP AlicloudNfsVolume status with ready condition").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if capacityChanged {
		return composed.UpdateStatus(state.ObjAsAlicloudNfsVolume()).
			SuccessErrorNil().
			ErrorLogMessage("Error updating SKR AlicloudNfsVolume status with Capacity change").
			SuccessLogMsg("Updated SKR AlicloudNfsVolume status with Capacity change").
			Run(ctx, state)
	}
	return nil, nil
}
