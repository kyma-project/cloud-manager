package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func deleteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if len(obj.Status.Id) == 0 {
		logger.Info("VpcPeering deleted before AWS peering is created", "vpcPeeringName", obj.Name)
		return nil, nil
	}

	for _, t := range state.routeTables {
		for _, r := range t.Routes {
			if ptr.Deref(r.VpcPeeringConnectionId, "xxx") == obj.Status.Id {

				err := state.client.DeleteRoute(ctx, t.RouteTableId, r.DestinationCidrBlock)

				lll := logger.WithValues(
					"vpcPeeringName", obj.Name,
					"vpcPeeringId", obj.Status.Id,
					"routeTableId", ptr.Deref(t.RouteTableId, "xxx"),
				)

				lll.Info("Deleting route")

				if err != nil {
					lll.Error(err, "Error deleting route")
				}
			}
		}
	}

	return nil, nil
}
