package iprange

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shouldAllocateIpRange(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return false
	}
	if len(state.ObjAsIpRange().Status.Cidr) > 0 {
		return false
	}
	if len(state.ObjAsIpRange().Spec.Cidr) > 0 {
		return false
	}
	return true
}

func allocateIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	logger := composed.LoggerFromCtx(ctx)

	cidr, err := iprangeallocate.AllocateCidr(iprangeallocate.DefaultMaskSize, state.existingCidrRanges)
	if err != nil {
		logger = logger.WithValues(
			"existingRanges", fmt.Sprintf("%v", state.existingCidrRanges),
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
