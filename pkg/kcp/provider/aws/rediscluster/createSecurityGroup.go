package rediscluster

import (
	"context"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func createSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisCluster()

	sgName := GetAwsElastiCacheSecurityGroupName(state.Obj().GetName())
	sgId, err := state.awsClient.CreateElastiCacheSecurityGroup(ctx, state.IpRange().Status.VpcId, sgName, []ec2types.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(sgName),
		},
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
		return awsmeta.LogErrorAndReturn(err, "Error creating security group", ctx)
	}

	logger = logger.WithValues("securityGroupId", sgId)
	logger.Info("Security group created")

	return composed.StopWithRequeue, nil
}
