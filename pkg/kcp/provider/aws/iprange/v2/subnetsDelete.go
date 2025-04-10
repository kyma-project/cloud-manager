package v2

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func subnetsDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.cloudResourceSubnets) == 0 {
		return nil, nil
	}

	anyDeleted := false
	for _, subnet := range state.cloudResourceSubnets {
		if subnet.State != ec2types.SubnetStateAvailable {
			continue
		}

		subnetId := ptr.Deref(subnet.SubnetId, "")

		lll := logger.WithValues("subnetId", subnetId)
		lll.Info("Deleting subnet")
		ccc := composed.LoggerIntoCtx(ctx, lll)

		err := state.awsClient.DeleteSubnet(ctx, subnetId)
		if x := awserrorhandling.HandleError(ccc, err, state, "KCP IpRange on delete subnet",
			cloudcontrolv1beta1.ReasonUnknown, "Error deleting AWS subnet"); x != nil {
			return x, nil
		}
		anyDeleted = true
	}

	if anyDeleted {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}
	return nil, nil
}
