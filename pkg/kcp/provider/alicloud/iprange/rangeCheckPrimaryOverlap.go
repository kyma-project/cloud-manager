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

// cidrOverlapsWith returns the first entry in forbidden that overlaps candidate,
// skipping unparseable entries and exact matches (idempotent re-association).
// Returns an empty string when no overlap is found.
func cidrOverlapsWith(candidate string, forbidden []string) (string, error) {
	rangeCidr, err := cidr.Parse(candidate)
	if err != nil {
		return "", fmt.Errorf("cannot parse CIDR %q: %w", candidate, err)
	}
	for _, other := range forbidden {
		if other == "" {
			continue
		}
		otherCidr, err := cidr.Parse(other)
		if err != nil {
			continue
		}
		if util.CidrEquals(rangeCidr.CIDR(), otherCidr.CIDR()) {
			continue
		}
		if util.CidrOverlap(rangeCidr.CIDR(), otherCidr.CIDR()) {
			return other, nil
		}
	}
	return "", nil
}

// rangeCheckPrimaryOverlap fails the IpRange if its CIDR overlaps the VPC primary
// CIDR or an existing secondary CIDR block, which AliCloud would reject on
// association (IllegalParam.SecondaryCidrBlock). Guards both the auto-allocated
// and the user-supplied spec.cidr paths.
func rangeCheckPrimaryOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	statusCidr := state.ObjAsIpRange().Status.Cidr
	forbidden := append([]string{state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR}, state.secondaryCidrBlocks...)

	overlapping, err := cidrOverlapsWith(statusCidr, forbidden)
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

	if overlapping != "" {
		logger.WithValues("cidr", statusCidr, "overlaps", overlapping).
			Info("AliCloud IpRange CIDR overlaps VPC address space")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
				Message: fmt.Sprintf("CIDR %s overlaps with VPC address range %s", statusCidr, overlapping),
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status due to VPC CIDR overlap").
			SuccessLogMsg("Forgetting AliCloud KCP IpRange due to VPC CIDR overlap").
			Run(ctx, state)
	}

	return nil, ctx
}
