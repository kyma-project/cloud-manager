package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
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
	tables := alicloudvpcpeeringclient.RouteTablesForStrategy(state.remoteRouteTables, strategy, shootName)
	for _, table := range tables {
		if err := state.remoteClient.DeleteRouteEntry(ctx, table.RouteTableId, localVpcCidr, instanceId); err != nil {
			return composed.LogErrorAndReturn(err, "Error deleting remote route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	return nil, ctx
}
