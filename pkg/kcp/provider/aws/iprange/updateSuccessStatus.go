package iprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateSuccessStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeError)
	meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  "True",
		Reason:  cloudresourcesv1beta1.ReasonReady,
		Message: "Additional IpRange(s) are provisioned",
	})
	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ReadyState

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	return nil, nil
}
