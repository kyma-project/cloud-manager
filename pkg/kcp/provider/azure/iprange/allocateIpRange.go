package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewAllocateIpRangeAction returns an Action that will populate state.ExistingCidrRanges
// with occupied cidr ranges so the allocation can pick a free slot.
func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		// atm we're considering only ranges defined in the Shoot, so we're not using
		// aws/iprange/state, but leaving extension possibility in the future to
		// create local state and acquire extended information from the AWS

		state := st.(iprangetypes.State)

		if len(state.Scope().Spec.Scope.Azure.Network.Nodes) == 0 {
			logger := composed.LoggerFromCtx(ctx)
			logger.Error(errors.New("network nodes empty"), "Azure scope has no nodes specified, unable to allocate IpRange cidr")
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
					Message: "Error due to unknown SKR nodes range",
				}).
				ErrorLogMessage("Failed patching KCP IpRange status with error due to unknown SKR nodes range").
				SuccessLogMsg("Forgetting KCP IpRange in error state due to unknown SKR nodes range").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, st)
		}

		state.SetExistingCidrRanges([]string{
			state.Scope().Spec.Scope.Azure.Network.Nodes,
			state.Scope().Spec.Scope.Azure.Network.Pods,
			state.Scope().Spec.Scope.Azure.Network.Services,
		})

		return nil, ctx
	}
}
