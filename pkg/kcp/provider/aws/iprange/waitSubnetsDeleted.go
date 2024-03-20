package iprange

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
	"time"
)

func waitSubnetsDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.cloudResourceSubnets) == 0 {
		return nil, nil
	}

	logger.
		WithValues(
			"waitingForSubnets",
			fmt.Sprintf("%v", pie.Map(state.cloudResourceSubnets, func(sn ec2Types.Subnet) string {
				return fmt.Sprintf(
					"%s/%s/%s/%s",
					pointer.StringDeref(sn.SubnetId, ""),
					pointer.StringDeref(sn.AvailabilityZone, ""),
					pointer.StringDeref(sn.CidrBlock, ""),
					sn.State,
				)
			})),
		).
		Info("Waiting for subnets to get deleted")

	return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
}
