package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func remoteRoutesDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		logger.Info("Skipping route deletion")
		return nil, ctx
	}

	if len(state.ObjAsVpcPeering().Status.RemoteId) == 0 {
		logger.Info("Skip deleting remote routes since VpcPeering.Status.RemoteId is empty")
		return nil, ctx
	}

	for _, t := range state.remoteRouteTables {

		shouldUpdateRouteTable := awsutil.ShouldUpdateRouteTable(t.Tags,
			state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy,
			state.Scope().Spec.ShootName)

		if !shouldUpdateRouteTable {
			continue
		}

		for _, r := range t.Routes {

			if ptr.Equal(r.VpcPeeringConnectionId, &state.ObjAsVpcPeering().Status.RemoteId) {
				err := state.remoteClient.DeleteRoute(ctx, t.RouteTableId, r.DestinationCidrBlock)

				lll := logger.WithValues(
					"remoteRouteTableId", ptr.Deref(t.RouteTableId, "xxx"),
					"destinationCidrBlock", ptr.Deref(r.DestinationCidrBlock, "xxx"),
				)

				if err != nil {
					if awsmeta.IsErrorRetryable(err) {
						return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
					}

					lll.Error(err, "Error deleting remote route")
					return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
				}

				lll.Info("Remote route deleted")
			}
		}
	}

	return nil, ctx
}
