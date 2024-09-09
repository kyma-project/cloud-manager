package redisinstance

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func createUserGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.userGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisInstance()

	out, err := state.awsClient.CreateUserGroup(ctx, GetAwsElastiCacheParameterGroupName(state.Obj().GetName()), []types.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: ptr.To(redisInstance.Spec.Scope.Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: ptr.To(state.Scope().Spec.ShootName),
		},
	})
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error creating user group", ctx)
	}

	logger = logger.WithValues("userGroupId", out.UserGroupId)
	logger.Info("User group created")

	return composed.StopWithRequeue, nil
}
