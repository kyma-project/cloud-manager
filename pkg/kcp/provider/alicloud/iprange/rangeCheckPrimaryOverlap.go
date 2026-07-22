package iprange

import (
	"context"
	"fmt"

	"github.com/3th1nk/cidr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// rangeCheckPrimaryOverlap fails the IpRange if its CIDR overlaps the VPC primary
// CIDR or an existing secondary CIDR block, which AliCloud would reject on
// association (IllegalParam.SecondaryCidrBlock). Guards both the auto-allocated
// and the user-supplied spec.cidr paths.
func rangeCheckPrimaryOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	rangeCidr, err := cidr.Parse(state.ObjAsIpRange().Status.Cidr)
	if err != nil {
		logger.Error(err, "Error parsing AliCloud IpRange CIDR for primary overlap check")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidCidr,
				Message: "Cannot parse CIDR",
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status with CIDR parse error").
			SuccessLogMsg("Forgetting AliCloud KCP IpRange with CIDR parse error").
			Run(ctx, state)
	}

	forbidden := []string{state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR}
	forbidden = append(forbidden, state.secondaryCidrBlocks...)

	for _, other := range forbidden {
		if other == "" {
			continue
		}
		otherCidr, err := cidr.Parse(other)
		if err != nil {
			continue
		}
		// An identical existing secondary block is idempotent, not an overlap error.
		if util.CidrEquals(rangeCidr.CIDR(), otherCidr.CIDR()) {
			continue
		}
		if util.CidrOverlap(rangeCidr.CIDR(), otherCidr.CIDR()) {
			logger.WithValues("cidr", state.ObjAsIpRange().Status.Cidr, "overlaps", other).
				Info("AliCloud IpRange CIDR overlaps VPC address space")
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
					Message: fmt.Sprintf("CIDR %s overlaps with VPC address range %s", state.ObjAsIpRange().Status.Cidr, other),
				}).
				ErrorLogMessage("Error patching AliCloud KCP IpRange status due to VPC CIDR overlap").
				SuccessLogMsg("Forgetting AliCloud KCP IpRange due to VPC CIDR overlap").
				Run(ctx, state)
		}
	}

	return nil, ctx
}
