package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewAllocateIpRangeAction reserves the VPC primary CIDR and the shoot's
// Pods/Services/Workers ranges so the shared allocateIpRange action picks a /22
// disjoint from the primary. AliCloud rejects a secondary CIDR block that overlaps
// the primary. Mirrors the AWS IpRange allocation flow.
func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(iprangetypes.State)

		if state.Scope().Spec.Scope.Alicloud == nil {
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
					Message: "AliCloud scope not populated",
				}).
				ErrorLogMessage("Failed patching KCP IpRange status with error due to missing AliCloud scope").
				SuccessLogMsg("Forgetting KCP IpRange in error state due to missing AliCloud scope").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, st)
		}

		vpcCidr := state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR
		if vpcCidr == "" {
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

		existing := []string{
			vpcCidr,
			state.Scope().Spec.Scope.Alicloud.Network.Pods,
			state.Scope().Spec.Scope.Alicloud.Network.Services,
		}
		for _, z := range state.Scope().Spec.Scope.Alicloud.Network.Zones {
			if z.Workers != "" {
				existing = append(existing, z.Workers)
			}
		}

		filtered := make([]string, 0, len(existing))
		for _, cidr := range existing {
			if cidr != "" {
				filtered = append(filtered, cidr)
			}
		}

		state.SetExistingCidrRanges(filtered)

		return nil, ctx
	}
}
