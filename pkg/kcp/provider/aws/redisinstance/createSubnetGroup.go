package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/ptr"
)

func createSubnetGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.subnetGroup != nil {
		return nil, ctx
	}

	redisInstance := state.ObjAsRedisInstance()

	logger := composed.LoggerFromCtx(ctx)
	subnetIds := pie.Map(state.IpRange().Status.Subnets, func(subnet cloudcontrolv1beta1.IpRangeSubnet) string {
		return subnet.Id
	})

	out, err := state.awsClient.CreateElastiCacheSubnetGroup(ctx, GetAwsElastiCacheSubnetGroupName(state.Obj().GetName()), subnetIds, []elasticachetypes.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: new(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: new(state.Name().String()),
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
		logger.Error(err, "Error creating subnet group")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create subnet group",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		updateErr := state.UpdateObjStatus(ctx)
		if updateErr != nil {
			return composed.LogErrorAndReturn(updateErr,
				"Error updating RedisInstance status due failed subnet group creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger = logger.WithValues("subnetGroupName", out.CacheSubnetGroup.CacheSubnetGroupName)
	logger.Info("Subnet group created")

	return composed.StopWithRequeue, nil
}
