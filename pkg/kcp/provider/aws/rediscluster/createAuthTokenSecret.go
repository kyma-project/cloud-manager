package rediscluster

import (
	"context"

	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"

	"k8s.io/utils/ptr"
)

func createAuthTokenSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.authTokenValue != nil {
		return nil, ctx
	}

	redisInstance := state.ObjAsRedisCluster()

	if !redisInstance.Spec.Instance.Aws.AuthEnabled {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)

	err := state.awsClient.CreateAuthTokenSecret(ctx, GetAwsAuthTokenSecretName(state.Obj().GetName()), []secretsmanagertypes.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
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
		return awsmeta.LogErrorAndReturn(err, "Error creating authToken secret", ctx)
	}

	logger.Info("AuthToken secret created")

	return composed.StopWithRequeue, nil
}
