package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkNupdateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//TBD: Check and see whether the desiredState == actualState
	match := true

	//If states do notmatch, continue to the next step
	if !match {
		return nil, nil
	}

	//If desiredState == actualState, update the status, and stop reconcilation
	meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeError)
	meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  "True",
		Reason:  cloudresourcesv1beta1.ReasonReady,
		Message: "IpRange(s) are provisioned",
	})
	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ReadyState

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	return composed.StopAndForget, nil
}
