package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	subnet := state.ObjAsGcpSubnet()

	hasReadyCondition := meta.FindStatusCondition(subnet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := subnet.Status.State == cloudcontrolv1beta1.StateReady

	if hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("Subnet status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	subnet.Status.State = cloudcontrolv1beta1.StateReady
	return composed.UpdateStatus(subnet).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Subnet is ready",
		}).
		ErrorLogMessage("Error updating KCP Subnet status after setting Ready condition").
		SuccessLogMsg("KCP Subnet is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
