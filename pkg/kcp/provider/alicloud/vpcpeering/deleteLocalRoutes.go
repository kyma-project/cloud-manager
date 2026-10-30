package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteLocalRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" || state.vpcPeering == nil {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	remoteCidrs := alicloudvpcpeeringclient.RemoteIpv4Cidrs(state.vpcPeering)
	for _, table := range state.routeTables {
		for _, cidr := range remoteCidrs {
			if err := state.client.DeleteRouteEntry(ctx, table.RouteTableId, cidr, instanceId); err != nil {
				return composed.LogErrorAndReturn(err, "Error deleting local route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	return nil, ctx
}
