package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering == nil {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	if awsutil.IsTerminated(state.vpcPeering) {
		logger.Info("VpcPeering can't be deleted at this stage",
			"peeringStatusCode", string(state.vpcPeering.Status.Code),
			"peeringStatusMessage", ptr.Deref(state.vpcPeering.Status.Message, ""))
		return nil, nil
	}

	logger.Info("Deleting VpcPeering")

	err := state.client.DeleteVpcPeeringConnection(ctx, state.vpcPeering.VpcPeeringConnectionId)

	if awsmeta.IsErrorRetryable(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	logger.Info("VpcPeering deleted")

	return nil, nil
}
