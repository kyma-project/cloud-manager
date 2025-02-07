package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func remoteRoutesDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		logger.Info("Skipping route deletion")
		return nil, nil
	}

	if state.remoteVpcPeering == nil {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	for _, t := range state.remoteRouteTables {

		for _, r := range t.Routes {

			if ptr.Equal(r.VpcPeeringConnectionId, state.remoteVpcPeering.VpcPeeringConnectionId) {
				err := state.remoteClient.DeleteRoute(ctx, t.RouteTableId, r.DestinationCidrBlock)

				lll := logger.WithValues(
					"routeTableId", ptr.Deref(t.RouteTableId, "xxx"),
					"destinationCidrBlock", ptr.Deref(r.DestinationCidrBlock, "xxx"),
				)

				if err != nil {
					if awsmeta.IsErrorRetryable(err) {
						return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
					}

					lll.Error(err, "Error deleting remote route")
					return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
				}

				lll.Info("Remote route deleted")
			}
		}
	}

	return nil, nil
}
