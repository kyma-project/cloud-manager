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

func rangeCheckVSwitchOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Load all vSwitches in the VPC to check for overlaps
	allVSwitches, err := state.client.DescribeVSwitchesByVpcId(ctx, state.vpcId)
	if err != nil {
		logger.Error(err, "Error loading AliCloud vSwitches for overlap check")
		return composed.StopWithRequeue, ctx
	}

	// Build set of CM-managed vSwitch CIDRs (those we're about to create or already created)
	cmCidrs := map[string]struct{}{}
	for _, vsw := range state.vSwitches {
		cmCidrs[vsw.CidrBlock] = struct{}{}
	}

	for _, zoneCidr := range state.zoneCidrs {
		rangeCidr, err := cidr.Parse(zoneCidr)
		if err != nil {
			continue
		}

		for _, vsw := range allVSwitches {
			// Skip our own vSwitches
			if _, isCM := cmCidrs[vsw.CidrBlock]; isCM {
				continue
			}

			vswCidr, err := cidr.Parse(vsw.CidrBlock)
			if err != nil {
				continue
			}

			if util.CidrOverlap(rangeCidr.CIDR(), vswCidr.CIDR()) {
				logger.WithValues(
					"range", zoneCidr,
					"vSwitchId", vsw.VSwitchId,
					"vSwitchCidr", vsw.CidrBlock,
				).Info("Range overlaps with existing vSwitch")

				state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
				return composed.PatchStatus(state.ObjAsIpRange()).
					SetExclusiveConditions(metav1.Condition{
						Type:    cloudcontrolv1beta1.ConditionTypeError,
						Status:  metav1.ConditionTrue,
						Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
						Message: fmt.Sprintf("Range %s overlaps with existing vSwitch %s with CIDR %s", zoneCidr, vsw.VSwitchId, vsw.CidrBlock),
					}).
					ErrorLogMessage("Error patching AliCloud KCP IpRange status due to vSwitch CIDR overlap").
					SuccessLogMsg("Forgetting AliCloud KCP IpRange due to vSwitch CIDR overlap").
					Run(ctx, state)
			}
		}
	}

	return nil, ctx
}
