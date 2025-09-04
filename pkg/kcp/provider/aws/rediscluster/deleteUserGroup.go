package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteUserGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.userGroup == nil {
		return nil, ctx
	}

	userGroupState := ptr.Deref(state.userGroup.Status, "")
	if userGroupState == awsmeta.ElastiCache_UserGroup_DELETING {
		return nil, ctx
	}

	logger.Info("Deleting userGroup")

	err := state.awsClient.DeleteUserGroup(ctx, *state.userGroup.UserGroupId)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting userGroup", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
