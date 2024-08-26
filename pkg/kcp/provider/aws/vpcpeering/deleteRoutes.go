package vpcpeering

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
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

		routeExists := pie.Any(t.Routes, func(r types.Route) bool {
			return ptr.Deref(r.VpcPeeringConnectionId, "xxx") == obj.Status.Id &&
				ptr.Equal(r.DestinationCidrBlock, state.remoteVpc.CidrBlock)
		})

		if routeExists {
			err := state.client.DeleteRoute(ctx, t.RouteTableId, state.remoteVpc.CidrBlock)

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

	return nil, nil
}
