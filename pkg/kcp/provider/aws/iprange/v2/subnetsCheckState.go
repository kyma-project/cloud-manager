package v2

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func subnetsCheckState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var info []string
	for _, subnet := range state.cloudResourceSubnets {
		info = append(info, fmt.Sprintf(
			"(%s-%s-%s-%s)",
			ptr.Deref(subnet.SubnetId, ""),
			ptr.Deref(subnet.AvailabilityZone, ""),
			ptr.Deref(subnet.CidrBlock, ""),
			subnet.State,
		))
	}

	logger := composed.LoggerFromCtx(ctx).
		WithValues("subnetsState", fmt.Sprintf("%v", info))

	for _, subnet := range state.cloudResourceSubnets {
		if subnet.State == ec2types.SubnetStatePending {
			logger.Info("Waiting KCP IpRange subnets to get ready state")
			return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
		}
	}

	return nil, nil
}
