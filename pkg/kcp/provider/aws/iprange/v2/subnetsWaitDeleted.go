package v2

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func subnetsWaitDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.cloudResourceSubnets) == 0 {
		return nil, nil
	}

	logger.
		WithValues(
			"waitingForSubnets",
			fmt.Sprintf("%v", pie.Map(state.cloudResourceSubnets, func(sn ec2types.Subnet) string {
				return fmt.Sprintf(
					"%s/%s/%s/%s",
					ptr.Deref(sn.SubnetId, ""),
					ptr.Deref(sn.AvailabilityZone, ""),
					ptr.Deref(sn.CidrBlock, ""),
					sn.State,
				)
			})),
		).
		Info("Waiting for subnets to get deleted")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
