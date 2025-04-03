package v3

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	ipRange := state.ObjAsIpRange()

	hasReadyCondition := meta.FindStatusCondition(ipRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := ipRange.Status.State == cloudcontrolv1beta1.StateReady

	if hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("IpRange status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	ipRange.Status.State = cloudcontrolv1beta1.StateReady
	return composed.UpdateStatus(ipRange).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "IpRange is ready",
		}).
		ErrorLogMessage("Error updating KCP IpRange status after setting Ready condition").
		SuccessLogMsg("KCP IpRange is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
