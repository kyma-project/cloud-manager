package v2

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func splitRangeByZones(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	ipRangeObj := state.ObjAsIpRange()

	if ipRangeObj.Status.Ranges != nil {
		return nil, nil
	}

	logger = logger.WithValues("cidr", ipRangeObj.Status.Cidr)

	wholeRange, err := cidr.Parse(ipRangeObj.Status.Cidr)
	if err != nil {
		logger.Error(err, "error parsing KCP IpRange CIDR")

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudresourcesv1beta1.ReasonInvalidCidr,
				Message: "Can not parse CIDR",
			}).
			ErrorLogMessage("Error patching KCP IpRange status with can not parse CIDR condition").
			SuccessLogMsg("Forgetting KCP IpRange with CIDR parse error").
			SuccessLogMsg("Forgetting KCP IpRange with invalid CIDR").
			Run(ctx, st)
	}

	numberOfSubnets := 1
	zoneCount := len(state.Scope().Spec.Scope.Aws.Network.Zones)
	for numberOfSubnets < zoneCount {
		numberOfSubnets = numberOfSubnets * 2
	}
	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, numberOfSubnets)
	if err != nil {
		logger.Error(err, "error splitting IpRange cidr")

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonCidrCanNotSplit,
				Message: fmt.Sprintf("Can not split CIDR to %d subnets", numberOfSubnets),
			}).
			ErrorLogMessage("Error patching KCP IpRange status after failed cidr splitting").
			SuccessLogMsg("Forgetting KCP IpRange after failed cidr splitting").
			Run(ctx, st)
	}
	subnetRanges = subnetRanges[:zoneCount]

	state.ObjAsIpRange().Status.Ranges = pie.Map(subnetRanges, func(c *cidr.CIDR) string {
		return c.CIDR().String()
	})

	logger.
		WithValues("ranges", state.ObjAsIpRange().Status.Ranges).
		Info("IpRange CIDR split")

	err = state.PatchObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP IpRange with split ranges", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
