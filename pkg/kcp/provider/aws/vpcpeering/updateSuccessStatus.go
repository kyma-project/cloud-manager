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

	obj := state.ObjAsVpcPeering()

	if len(obj.Status.Id) > 0 &&
		len(obj.Status.RemoteId) > 0 &&
		meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrol1beta1.ConditionTypeReady) {
		// all already set and saved
		return nil, nil
	}

	obj.Status.State = string(state.vpcPeering.Status.Code)

	return composed.UpdateStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrol1beta1.ConditionTypeReady,
			Status:  "True",
			Reason:  cloudcontrol1beta1.ReasonReady,
			Message: "Additional VpcPeerings(s) are provisioned",
		}).
		ErrorLogMessage("Error updating VpcPeering success status after setting Ready condition").
		SuccessLogMsg("KPC VpcPeering is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
