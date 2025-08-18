package rediscluster

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"

	"k8s.io/utils/ptr"
)

func createSubnetGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.subnetGroup != nil {
		return nil, ctx
	}

	redisInstance := state.ObjAsRedisCluster()

	logger := composed.LoggerFromCtx(ctx)
	subnetIds := pie.Map(state.IpRange().Status.Subnets, func(subnet cloudcontrolv1beta1.IpRangeSubnet) string {
		return subnet.Id
	})

	out, err := state.awsClient.CreateElastiCacheSubnetGroup(ctx, GetAwsElastiCacheSubnetGroupName(state.Obj().GetName()), subnetIds, []elasticachetypes.Tag{
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
		return awsmeta.LogErrorAndReturn(err, "Error creating subnet group", ctx)
	}

	logger = logger.WithValues("subnetGroupName", out.CacheSubnetGroup.CacheSubnetGroupName)
	logger.Info("Subnet group created")

	return composed.StopWithRequeue, nil
}
