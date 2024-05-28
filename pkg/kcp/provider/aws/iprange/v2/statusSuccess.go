package v2

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusSuccess(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ReadyState
	return composed.PatchStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonReady,
			Message: "Additional IpRange(s) are provisioned",
		}).
		ErrorLogMessage("Error patching KCP IpRange status with ready state").
		SuccessLogMsg("Forgetting KCP IpRange with ready state").
		Run(ctx, state)
}
