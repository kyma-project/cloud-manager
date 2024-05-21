package v1

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkSubnetOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	crSubnets := map[string]struct{}{}
	for _, crs := range state.cloudResourceSubnets {
		crSubnets[*crs.SubnetId] = struct{}{}
	}

	for _, subnet := range state.allSubnets {
		_, isCloudResourcesSubnet := crSubnets[*subnet.SubnetId]
		if isCloudResourcesSubnet {
			continue
		}

		for _, r := range state.ObjAsIpRange().Status.Ranges {
			rangeCidr, err := cidr.Parse(r)
			if err != nil {
				continue
			}
			subnetCidr, err := cidr.Parse(*subnet.CidrBlock)
			if err != nil {
				continue
			}

			if util.CidrOverlap(rangeCidr.CIDR(), subnetCidr.CIDR()) {
				logger.WithValues(
					"range", r,
					"subnetId", *subnet.SubnetId,
					"subnetCidr", *subnet.CidrBlock,
				).Info("Range overlaps with existing subnet")

				return composed.UpdateStatus(state.ObjAsIpRange()).
					SetExclusiveConditions(metav1.Condition{
						Type:    cloudcontrolv1beta1.ConditionTypeError,
						Status:  metav1.ConditionTrue,
						Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
						Message: fmt.Sprintf("Range %s overlaps with existing subnet %s with cidr %s", r, *subnet.SubnetId, *subnet.CidrBlock),
					}).
					ErrorLogMessage("Error updating KCP IpRange status with error condition due to cidr overlaps with existing subnet").
					SuccessError(composed.StopAndForget).
					Run(ctx, state)
			}
		}
	}

	return nil, nil
}
