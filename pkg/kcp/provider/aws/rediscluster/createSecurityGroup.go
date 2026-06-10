package rediscluster

import (
	"context"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisCluster()

	sgName := GetAwsElastiCacheSecurityGroupName(state.Obj().GetName())
	sgId, err := state.awsClient.CreateElastiCacheSecurityGroup(ctx, state.IpRange().Status.VpcId, sgName, []ec2types.Tag{
		{
			Key:   new("Name"),
			Value: new(sgName),
		},
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: new(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: new(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: new(redisInstance.Spec.Scope.Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: new(state.Scope().Spec.ShootName),
		},
	})
	if err != nil {
		logger.Error(err, "Error creating security group")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create security group",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		updateErr := state.UpdateObjStatus(ctx)
		if updateErr != nil {
			return composed.LogErrorAndReturn(updateErr,
				"Error updating RedisCluster status due failed security group creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger = logger.WithValues("securityGroupId", sgId)
	logger.Info("Security group created")

	return composed.StopWithRequeue, nil
}
