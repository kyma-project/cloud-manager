package vpcpeering

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("GCP VPC Peering Update Status")

	if composed.MarkedForDeletionPredicate(ctx, state) {
		logger.Info("GCP VPC Peering is marked for deletion")
		return nil, nil
	}

	if meta.IsStatusConditionTrue(
		*state.ObjAsVpcPeering().Conditions(),
		cloudcontrol1beta1.ConditionTypeReady,
	) {
		return nil, nil
	}

	return composed.UpdateStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrol1beta1.ConditionTypeReady,
			Status:  "True",
			Reason:  cloudcontrol1beta1.ReasonReady,
			Message: "VpcPeering :" + state.remotePeeringName + " is provisioned",
		}).
		ErrorLogMessage("Error updating VpcPeering success status after setting Ready condition").
		SuccessLogMsg("KPC VpcPeering is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
