package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.securityGroup == nil {
		return nil, ctx
	}

	sgId := ptr.Deref(state.securityGroup.GroupId, "")
	logger.WithValues("securityGroupId", sgId).Info("Deleting security group")

	err := state.awsClient.DeleteElastiCacheSecurityGroup(ctx, sgId)
	if err != nil {
		// ENI from ElastiCache cluster may not yet be released — requeue silently
		if awsmeta.IsDependencyViolation(err) {
			logger.WithValues("securityGroupId", sgId).Info("Security group has dependent object, will retry")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}
		return awsmeta.LogErrorAndReturn(err, "Error deleting security group", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
