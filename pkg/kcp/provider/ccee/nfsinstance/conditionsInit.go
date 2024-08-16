package nfsinstance

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func conditionsInit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false
	if state.ObjAsNfsInstance().Status.State == "" {
		changed = true
		state.ObjAsNfsInstance().Status.State = cloudcontrol1beta1.ProcessingState
	}
	if len(state.ObjAsNfsInstance().Status.Conditions) == 0 {
		changed = true
		state.ObjAsNfsInstance().Status.Conditions = []metav1.Condition{
			{
				Type:    ConditionTypeNetworkFound,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
			{
				Type:    ConditionTypeShareNetworkCreated,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
			{
				Type:    ConditionTypeShareCreated,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
			{
				Type:    ConditionTypeAccessGranted,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
			{
				Type:    ConditionTypeAvailable,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
			{
				Type:    ConditionTypeEndpointsRead,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonPending,
				Message: ConditionReasonPending,
			},
		}
	}

	if changed {
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating initial conditions").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, nil
}
