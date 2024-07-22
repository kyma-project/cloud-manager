package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteSubnetGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnetGroup == nil {
		return nil, nil
	}

	logger.
		WithValues("subnetGroupName", ptr.Deref(state.subnetGroup.CacheSubnetGroupName, "")).
		Info("Deleting subnet group")

	err := state.awsClient.DeleteElastiCacheSubnetGroup(ctx, ptr.Deref(state.subnetGroup.CacheSubnetGroupName, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting subnet group", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
