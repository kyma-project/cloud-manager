package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func remoteRoutesDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}
	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		return nil, ctx
	}

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" {
		return nil, ctx
	}

	// localVpcCidr is the route destination; shootName is the tag key for MATCHED/UNMATCHED strategy.
	if state.Scope().Spec.Scope.Alicloud == nil {
		return nil, ctx
	}
	localVpcCidr := state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR
	if localVpcCidr == "" {
		return nil, ctx
	}
	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := routeTablesForStrategy(state.remoteRouteTables, strategy, shootName)
	for _, table := range tables {
		if err := state.remoteClient.DeleteRouteEntry(ctx, table.RouteTableId, localVpcCidr, instanceId); err != nil {
			return composed.LogErrorAndReturn(err, "Error deleting remote route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	return nil, ctx
}

func deleteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" || state.vpcPeering == nil {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := routeTablesForStrategy(state.routeTables, strategy, shootName)
	for _, table := range tables {
		for _, cidr := range state.vpcPeering.RemoteIpv4Cidrs {
			if err := state.client.DeleteRouteEntry(ctx, table.RouteTableId, cidr, instanceId); err != nil {
				return composed.LogErrorAndReturn(err, "Error deleting local route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	return nil, ctx
}

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" {
		return nil, ctx
	}

	if err := state.client.DeleteVpcPeerConnection(ctx, instanceId); err != nil {
		if alicloudvpcpeeringclient.IsRetryable(err) {
			// IncorrectStatus fires when the connection is already transitioning; safe to retry.
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}
		return composed.LogErrorAndReturn(err, "Error deleting AliCloud VpcPeerConnection", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, ctx
}
