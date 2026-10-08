package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitRedisDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	instanceId := state.ObjAsRedisCluster().Status.Id
	if instanceId == "" {
		return nil, ctx
	}
	info, err := state.client.DescribeInstance(ctx, instanceId)
	if err != nil {
		if alicloudclient.IsNotFoundErr(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err,
			"Error describing AliCloud r-kvstore cluster instance during delete wait",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}
	if info == nil || info.InstanceStatus == alicloudclient.InstanceStatusReleased {
		return nil, ctx
	}
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
