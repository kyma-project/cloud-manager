package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering.Status.Code != ec2types.VpcPeeringConnectionStateReasonCodeActive {
		logger.Info("Waiting for peering Connected state",
			"Id", *state.vpcPeering.VpcPeeringConnectionId,
			"PeeringStatus", state.vpcPeering.Status.Code)

		changed := false

		if state.ObjAsVpcPeering().Status.State != string(state.vpcPeering.Status.Code) {
			state.ObjAsVpcPeering().Status.State = string(state.vpcPeering.Status.Code)
			changed = true
		}

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeError) {
			changed = true
		}
		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if changed {
			return composed.PatchStatus(state.ObjAsVpcPeering()).
				ErrorLogMessage("Error patching KCP VpcPeering status on waiting peering active").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
				Run(ctx, state)
		}
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	return nil, nil
}
