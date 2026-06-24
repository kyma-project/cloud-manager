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

// NewAllocateIpRangeAction returns an Action that populates state.ExistingCidrRanges
// with occupied CIDR ranges from the AliCloud Scope so the common allocator can pick a free slot.
// The VPC CIDR is placed first so the allocator starts within the VPC range.
func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(iprangetypes.State)

		if len(state.Scope().Spec.Scope.Alicloud.Network.Nodes) == 0 {
			logger := composed.LoggerFromCtx(ctx)
			logger.Error(errors.New("network nodes empty"), "AliCloud scope has no nodes specified, unable to allocate IpRange cidr")
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

		vpcCidr := state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR
		if vpcCidr == "" {
			logger := composed.LoggerFromCtx(ctx)
			logger.Error(errors.New("vpc cidr empty"), "AliCloud scope has no VPC CIDR specified, unable to allocate IpRange cidr")
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
					Message: "Error due to unknown VPC CIDR range",
				}).
				ErrorLogMessage("Failed patching KCP IpRange status with error due to unknown VPC CIDR range").
				SuccessLogMsg("Forgetting KCP IpRange in error state due to unknown VPC CIDR range").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, st)
		}

		// VPC CIDR is placed first so the common allocator starts within the VPC range.
		// Nodes/Pods/Services follow to prevent overlap with shoot networking.
		state.SetExistingCidrRanges([]string{
			vpcCidr,
			state.Scope().Spec.Scope.Alicloud.Network.Nodes,
			state.Scope().Spec.Scope.Alicloud.Network.Pods,
			state.Scope().Spec.Scope.Alicloud.Network.Services,
		})

		return nil, ctx
	}
}
