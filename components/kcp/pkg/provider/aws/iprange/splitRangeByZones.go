package iprange

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func splitRangeByZones(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	ipRangeObj := state.ObjAsIpRange()

	if ipRangeObj.Status.Ranges != nil {
		return nil, nil
	}

	wholeRange, err := cidr.Parse(ipRangeObj.Spec.Cidr)
	if err != nil {
		err = fmt.Errorf("error parsing IpRange CIDR: %w", err)
		logger.Error(err, "Error splitting IpRange by zones")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonInvalidCidr,
			Message: "Can not parse CIDR",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			logger.Error(err, "error updating IpRange status condition")
			return composed.StopWithRequeue, nil
		}
		return composed.StopAndForget, nil
	}

	numberOfSubnets := 1
	zoneCount := len(state.Scope().Spec.Scope.Aws.Network.Zones)
	for numberOfSubnets < zoneCount {
		numberOfSubnets = numberOfSubnets * 2
	}
	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, numberOfSubnets)
	if err != nil {
		err = fmt.Errorf("error splitting IpRange cidr: %w", err)
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonCidrCanNotSplit,
			Message: fmt.Sprintf("Can not split CIDR to %d subnets", numberOfSubnets),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			logger.Error(err, "error updating IpRange status condition")
			return composed.StopWithRequeue, nil
		}
		return composed.StopAndForget, nil
	}
	subnetRanges = subnetRanges[:zoneCount]

	state.ObjAsIpRange().Status.Ranges = pie.Map(subnetRanges, func(c *cidr.CIDR) string {
		return c.CIDR().String()
	})

	logger.
		WithValues("ranges", state.ObjAsIpRange().Status.Ranges).
		Info("IpRange CIDR split")

	err = state.UpdateObjStatus(ctx)
	if err != nil {
		err = fmt.Errorf("error updating IpRange status: %w", err)
		return composed.LogErrorAndReturn(err, "Error splitting IpRange CIDR", composed.StopWithRequeue, nil)
	}

	return composed.StopWithRequeue, nil
}
