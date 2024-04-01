package vpcpeering

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateSuccessStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrol1beta1.ConditionTypeError)
	meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
		Type:    cloudcontrol1beta1.ConditionTypeReady,
		Status:  "True",
		Reason:  cloudcontrol1beta1.ReasonReady,
		Message: "Additional VpcPeerings(s) are provisioned",
	})
	state.ObjAsVpcPeering().Status.State = cloudcontrol1beta1.ReadyState

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VpcPeering success status", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
