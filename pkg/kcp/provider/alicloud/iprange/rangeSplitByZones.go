package iprange

import (
	"context"
	"fmt"

	"github.com/3th1nk/cidr"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func rangeSplitByZones(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsIpRange().Status.Ranges) > 0 {
		state.zoneCidrs = state.ObjAsIpRange().Status.Ranges
		return nil, ctx
	}

	zoneCount := len(state.Scope().Spec.Scope.Alicloud.Network.Zones)
	if zoneCount == 0 {
		zoneCount = 1
	}

	wholeRange, err := cidr.Parse(state.ObjAsIpRange().Status.Cidr)
	if err != nil {
		logger.Error(err, "Error parsing AliCloud IpRange CIDR")
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

	// Round up to next power of 2 to satisfy subnetting requirements
	numberOfSubnets := 1
	for numberOfSubnets < zoneCount {
		numberOfSubnets *= 2
	}

	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, numberOfSubnets)
	if err != nil {
		logger.Error(err, "Error splitting AliCloud IpRange CIDR by zones")
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCidrCanNotSplit,
				Message: fmt.Sprintf("Cannot split CIDR to %d subnets", numberOfSubnets),
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status after failed CIDR splitting").
			SuccessLogMsg("Forgetting AliCloud KCP IpRange after failed CIDR splitting").
			Run(ctx, state)
	}

	// Take only as many subnets as there are zones
	subnetRanges = subnetRanges[:zoneCount]

	state.ObjAsIpRange().Status.Ranges = pie.Map(subnetRanges, func(c *cidr.CIDR) string {
		return c.CIDR().String()
	})
	state.zoneCidrs = state.ObjAsIpRange().Status.Ranges

	logger.WithValues("ranges", state.zoneCidrs).Info("AliCloud IpRange CIDR split by zones")

	return composed.PatchStatus(state.ObjAsIpRange()).
		ErrorLogMessage("Error patching AliCloud KCP IpRange status with zone ranges").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
