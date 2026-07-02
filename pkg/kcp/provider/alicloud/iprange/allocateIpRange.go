package iprange

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewAllocateIpRangeAction allocates a /22 CIDR within the shoot VPC CIDR,
// avoiding overlap with the shoot's Nodes/Pods/Services ranges.
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

		as, err := iprangeallocate.NewAddressSpace(vpcCidr)
		if err != nil {
			return composed.LogErrorAndReturn(
				fmt.Errorf("error creating address space from VPC CIDR %s: %w", vpcCidr, err),
				"Error creating AliCloud IpRange address space",
				composed.StopWithRequeue,
				ctx,
			)
		}

		occupied := []string{
			state.Scope().Spec.Scope.Alicloud.Network.Pods,
			state.Scope().Spec.Scope.Alicloud.Network.Services,
		}
		for _, z := range state.Scope().Spec.Scope.Alicloud.Network.Zones {
			if z.Workers != "" {
				occupied = append(occupied, z.Workers)
			}
		}
		for _, cidr := range occupied {
			if cidr == "" {
				continue
			}
			// ignore out-of-space-range ranges (nodes/pods/services may be outside VPC CIDR)
			_ = as.Reserve(cidr)
		}

		cidr, err := as.Allocate(iprangeallocate.DefaultMaskSize)
		if err != nil {
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
					Message: fmt.Sprintf("Unable to allocate CIDR within VPC %s", vpcCidr),
				}).
				ErrorLogMessage("Error patching KCP IpRange status after failed cidr allocation").
				SuccessLogMsg("Forgetting KCP IpRange with failed CIDR allocation").
				Run(ctx, st)
		}

		state.ObjAsIpRange().Status.Cidr = cidr

		return composed.PatchStatus(state.ObjAsIpRange()).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, st)
	}
}
