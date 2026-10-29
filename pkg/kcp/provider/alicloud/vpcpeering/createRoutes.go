package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	remoteCidrs := alicloudvpcpeeringclient.RemoteIpv4Cidrs(state.vpcPeering)
	if len(remoteCidrs) == 0 {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	// Use the shoot name (not the CIDR) as the tag key for MATCHED/UNMATCHED filtering,
	// consistent with the AWS pattern (ShouldUpdateRouteTable uses shoot name as tag key).
	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := alicloudvpcpeeringclient.RouteTablesForStrategy(state.routeTables, strategy, shootName)

	instanceId := state.ObjAsVpcPeering().Status.Id
	for _, table := range tables {
		for _, cidr := range remoteCidrs {
			if err := state.client.CreateRouteEntry(ctx, table.RouteTableId, cidr, instanceId); err != nil {
				return composed.LogErrorAndReturn(err, "Error creating local route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	return nil, ctx
}
