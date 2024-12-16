package v2

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		// atm we're considering only ranges defined in the Shoot, so we're not using
		// aws/iprange/state, but leaving extension possibility in the future to
		// create local state and acquire extended information from the AWS

		state := st.(iprangetypes.State)

		if len(state.Scope().Spec.Scope.Aws.Network.Nodes) == 0 {
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
				Run(ctx, st)
		}

		state.SetExistingCidrRanges([]string{
			state.Scope().Spec.Scope.Aws.Network.Nodes,
			state.Scope().Spec.Scope.Aws.Network.Pods,
			state.Scope().Spec.Scope.Aws.Network.Services,
		})

		return nil, ctx
	}
}
