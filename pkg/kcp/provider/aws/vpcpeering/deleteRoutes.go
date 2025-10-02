package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsVpcPeering().Status.Id) == 0 {
		logger.Info("Skip deleting local routes since VpcPeering.Status.Id is empty")
		return nil, ctx
	}

	for _, t := range state.routeTables {
		for _, r := range t.Routes {
			if ptr.Equal(r.VpcPeeringConnectionId, &state.ObjAsVpcPeering().Status.Id) {

				err := state.client.DeleteRoute(ctx, t.RouteTableId, r.DestinationCidrBlock)

				if awsmeta.IsErrorRetryable(err) {
					return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
				}

				lll := logger.WithValues(
					"routeTableId", ptr.Deref(t.RouteTableId, "xxx"),
					"destinationCidrBlock", ptr.Deref(r.DestinationCidrBlock, "xxx"),
				)

				if err != nil {
					lll.Error(err, "Error deleting route")
					return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
				}

				lll.Info("Route deleted")
			}
		}
	}

	return nil, ctx
}
