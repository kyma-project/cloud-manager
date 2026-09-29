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

// routeTablesForStrategy returns the subset of route tables to act on for the given strategy.
// For MATCHED/UNMATCHED the shootName parameter is used as the tag KEY: a table is "matched"
// when it carries that key with a non-empty value (e.g. tag key = "<shoot-name>", value = "true").
// AUTO selects all tables; NONE skips all.
func routeTablesForStrategy(tables []alicloudvpcpeeringclient.RouteTableInfo, strategy, shootName string) []alicloudvpcpeeringclient.RouteTableInfo {
	switch strategy {
	case "NONE":
		return nil
	case "MATCHED":
		var result []alicloudvpcpeeringclient.RouteTableInfo
		for _, t := range tables {
			if t.Tags[shootName] != "" {
				result = append(result, t)
			}
		}
		return result
	case "UNMATCHED":
		var result []alicloudvpcpeeringclient.RouteTableInfo
		for _, t := range tables {
			if t.Tags[shootName] == "" {
				result = append(result, t)
			}
		}
		return result
	default: // AUTO
		return tables
	}
}
