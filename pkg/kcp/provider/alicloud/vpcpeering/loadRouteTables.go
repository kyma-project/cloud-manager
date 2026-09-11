package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.localVpcId == "" {
		return nil, ctx
	}

	tables, err := state.client.DescribeRouteTables(ctx, state.localVpcId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading local route tables", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.routeTables = tables

	return nil, ctx
}

func loadRemoteRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	remoteNet := state.RemoteNetwork()
	if remoteNet.Status.Network == nil || remoteNet.Status.Network.Alicloud == nil {
		return nil, ctx
	}

	remoteVpcId := remoteNet.Status.Network.Alicloud.VpcId
	tables, err := state.remoteClient.DescribeRouteTables(ctx, remoteVpcId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading remote route tables", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.remoteRouteTables = tables

	return nil, ctx
}

func routeTablesForStrategy(tables []alicloudvpcpeeringclient.RouteTableInfo, strategy, localNetworkTag string) []alicloudvpcpeeringclient.RouteTableInfo {
	switch strategy {
	case "NONE":
		return nil
	case "MATCHED":
		var result []alicloudvpcpeeringclient.RouteTableInfo
		for _, t := range tables {
			if t.Tags[localNetworkTag] != "" {
				result = append(result, t)
			}
		}
		return result
	case "UNMATCHED":
		var result []alicloudvpcpeeringclient.RouteTableInfo
		for _, t := range tables {
			if t.Tags[localNetworkTag] == "" {
				result = append(result, t)
			}
		}
		return result
	default: // AUTO
		return tables
	}
}
