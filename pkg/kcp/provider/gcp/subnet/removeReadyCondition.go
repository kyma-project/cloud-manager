package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	subnet := state.ObjAsGcpSubnet()

	readyCond := meta.FindStatusCondition(*subnet.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		return nil, nil
	}

	logger.Info("Removing Ready condition")

	meta.RemoveStatusCondition(subnet.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	subnet.Status.State = cloudcontrolv1beta1.StateDeleting
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Subnet status after removing Ready condition", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
