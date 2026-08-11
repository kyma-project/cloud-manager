package iprange

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vSwitchDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.vSwitches) == 0 {
		return nil, ctx
	}

	for _, vsw := range state.vSwitches {
		logger.Info("Deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)

		err := state.client.DeleteVSwitch(ctx, vsw.VSwitchId)
		if err != nil {
			logger.Error(err, "Error deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: fmt.Sprintf("Error deleting VSwitch %s; see controller logs for details", vsw.VSwitchId),
				}).
				ErrorLogMessage("Error patching AliCloud KCP IpRange status after failed VSwitch delete").
				SuccessError(composed.StopWithRequeue).
				Run(ctx, state)
		}
	}

	// Requeue to let AliCloud finish removing vSwitches before attempting
	// to disassociate the VPC address space (which fails if vSwitches still exist).
	return composed.StopWithRequeue, ctx
}
