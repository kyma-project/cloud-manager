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
	// Defensive guard against a fresh-create → downgrade race: if the transient
	// user group was just created (predicate fired) and hasn't reached ACTIVE
	// yet, AWS rejects DeleteUserGroup on a CREATING resource and lands the
	// cluster in an error state. Wait for the status to settle before
	// attempting the delete.
	if userGroupState == awsmeta.ElastiCache_UserGroup_CREATING {
		logger.Info("User group is still CREATING, requeueing before delete")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	logger.Info("Deleting userGroup")

	err := state.awsClient.DeleteUserGroup(ctx, *state.userGroup.UserGroupId)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting userGroup", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
