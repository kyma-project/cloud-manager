package vpcpeering

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/ptr"
)

func peeringLocalWaitReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.localPeering == nil {
		return composed.StopWithRequeue, nil
	}

	if ptr.Deref(state.localPeering.Properties.PeeringState, "") != armnetwork.VirtualNetworkPeeringStateConnected {
		logger.Info("Waiting for peering Connected state",
			"azureLocalPeeringId", ptr.Deref(state.localPeering.ID, ""),
			"azurePeeringState", ptr.Deref(state.localPeering.Properties.PeeringState, ""))

		changed := false

		if state.ObjAsVpcPeering().Status.State != cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated {
			state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated
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
				ErrorLogMessage("Error patching KCP VpcPeering status when state initiated on wait peering connected").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
				Run(ctx, state)
		}

		// azure peering connects quickly so we're aggressive and requeue with small delay
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	return nil, nil
}
