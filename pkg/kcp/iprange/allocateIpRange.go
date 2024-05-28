package iprange

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func allocateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsIpRange().Status.Cidr) > 0 {
		// already allocated
		return nil, nil
	}
	if len(state.ObjAsIpRange().Spec.Cidr) > 0 {
		// already allocated
		return nil, nil
	}

	if len(state.Scope().Spec.Scope.Aws.Network.Nodes) == 0 {
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ErrorState
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

	existingRanges := []string{
		state.Scope().Spec.Scope.Aws.Network.Nodes,
		state.Scope().Spec.Scope.Aws.Network.Pods,
		state.Scope().Spec.Scope.Aws.Network.Services,
	}

	cidr, err := iprangeallocate.AllocateCidr(22, existingRanges)
	if err != nil {
		logger = logger.WithValues(
			"existingRanges", fmt.Sprintf("%v", existingRanges),
		)
		ctx = composed.LoggerIntoCtx(ctx, logger)
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCidrAllocationFailed,
				Message: "Unable to allocate CIDR",
			}).
			ErrorLogMessage("Error patching KCP IpRange status after failed cidr allocation").
			SuccessLogMsg("Forgetting KCP IpRange with failed CIDR allocation").
			Run(ctx, st)
	}

	state.ObjAsIpRange().Status.Cidr = cidr

	return composed.PatchStatus(state.ObjAsIpRange()).
		SuccessErrorNil().
		Run(ctx, state)
}
