package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func waitUserGroupActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.userGroup == nil {
		return nil, nil
	}

	userGroupState := ptr.Deref(state.userGroup.Status, "")
	if userGroupState == awsmeta.ElastiCache_UserGroup_ACTIVE {
		return nil, nil
	}

	logger.Info("Redis elasticache user group is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
