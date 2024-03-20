package iprange

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func deleteSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.cloudResourceSubnets) == 0 {
		return nil, nil
	}

	for _, subnet := range state.cloudResourceSubnets {
		if subnet.State != ec2Types.SubnetStateAvailable {
			continue
		}

		subnetId := pointer.StringDeref(subnet.SubnetId, "")

		lll := logger.WithValues("subnetId", subnetId)
		lll.Info("Deleting subnet")

		err := state.client.DeleteSubnet(ctx, subnetId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error deleting subnet", composed.StopWithRequeue, composed.LoggerIntoCtx(ctx, lll))
		}
	}

	return nil, nil
}
