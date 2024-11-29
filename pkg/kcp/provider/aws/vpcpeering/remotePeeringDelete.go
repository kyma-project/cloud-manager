package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func remotePeeringDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		return nil, nil
	}
	if state.remoteVpcPeering == nil {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	if awsutil.IsTerminated(state.remoteVpcPeering) {
		logger.Info("Remote VpcPeering can't be deleted at this stage",
			"peeringStatusCode", string(state.remoteVpcPeering.Status.Code),
			"peeringStatusMessage", ptr.Deref(state.remoteVpcPeering.Status.Message, ""))
		return nil, nil
	}

	logger.Info("Deleting remote VpcPeering")

	err := state.remoteClient.DeleteVpcPeeringConnection(ctx, state.remoteVpcPeering.VpcPeeringConnectionId)

	if err != nil {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err,
				"Error deleting AWS VPC peering connection but skipping as marked for deletion",
				nil,
				ctx)
		}

		if awsmeta.IsErrorRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}
	}

	logger.Info("Remote VpcPeering deleted")

	return nil, nil
}
