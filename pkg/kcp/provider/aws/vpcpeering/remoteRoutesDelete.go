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
		return nil, nil
	}

	if len(state.ObjAsVpcPeering().Status.RemoteId) == 0 {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	for _, t := range state.remoteRouteTables {
		for _, r := range t.Routes {
			if ptr.Deref(r.VpcPeeringConnectionId, "xxx") == state.ObjAsVpcPeering().Status.RemoteId {

				err := state.remoteClient.DeleteRoute(ctx, t.RouteTableId, r.DestinationCidrBlock)

				if awsmeta.IsErrorRetryable(err) {
					return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
				}

				lll := logger.WithValues(
					"routeTableId", ptr.Deref(t.RouteTableId, "xxx"),
					"destinationCidrBlock", ptr.Deref(r.DestinationCidrBlock, "xxx"),
				)

				if err != nil {
					lll.Error(err, "Error deleting remote route")
					return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
				}

				lll.Info("Remote route deleted")
			}
		}
	}

	return nil, nil
}
