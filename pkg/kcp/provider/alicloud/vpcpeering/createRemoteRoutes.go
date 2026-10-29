package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// The destination CIDR for remote routes pointing at the local (Kyma shoot) VPC.
	// Scope.Scope.Alicloud.VpcNetwork is the VPC name, not the CIDR.
	// The actual CIDR is in Scope.Scope.Alicloud.Network.VPC.CIDR.
	localVpcCidr := state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR
	if localVpcCidr == "" {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := alicloudvpcpeeringclient.RouteTablesForStrategy(state.remoteRouteTables, strategy, shootName)

	instanceId := state.ObjAsVpcPeering().Status.Id
	for _, table := range tables {
		if err := state.remoteClient.CreateRouteEntry(ctx, table.RouteTableId, localVpcCidr, instanceId); err != nil {
			return composed.LogErrorAndReturn(err, "Error creating remote route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	return nil, ctx
}
