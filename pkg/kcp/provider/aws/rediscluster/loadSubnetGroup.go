package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadSubnetGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.subnetGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeElastiCacheSubnetGroup(ctx, GetAwsElastiCacheSubnetGroupName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing subnet groups", ctx)
	}

	if len(list) > 0 {
		state.subnetGroup = &list[0]
		logger = logger.WithValues("subnetGroupName", ptr.Deref(state.subnetGroup.CacheSubnetGroupName, ""))
		logger.Info("ElastiCache subnet group found and loaded")
		return nil, composed.LoggerIntoCtx(ctx, logger)
	}

	logger.Info("ElastiCache subnet group not found")

	return nil, nil
}
