package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadUserGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.userGroup != nil {
		return nil, ctx
	}

	userGroup, err := state.awsClient.DescribeUserGroup(ctx, GetAwsElastiCacheUserGroupName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error getting user group", ctx)
	}

	if userGroup == nil {
		return nil, ctx
	}

	state.userGroup = userGroup

	return nil, ctx
}
