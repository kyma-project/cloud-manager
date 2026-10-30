package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createLocalRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	remoteCidrs := alicloudvpcpeeringclient.RemoteIpv4Cidrs(state.vpcPeering)
	if len(remoteCidrs) == 0 {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	instanceId := state.ObjAsVpcPeering().Status.Id
	for _, table := range state.routeTables {
		for _, cidr := range remoteCidrs {
			if err := state.client.CreateRouteEntry(ctx, table.RouteTableId, cidr, instanceId); err != nil {
				return composed.LogErrorAndReturn(err, "Error creating local route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	return nil, ctx
}
