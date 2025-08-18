package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"k8s.io/utils/ptr"
)

func createParameterGroup(getParamGroup func(*State) *elasticachetypes.CacheParameterGroup, name string) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)

		if getParamGroup(state) != nil {
			return nil, ctx
		}

		redisInstance := state.ObjAsRedisInstance()

		logger := composed.LoggerFromCtx(ctx)

		family := GetAwsElastiCacheParameterGroupFamily(redisInstance.Spec.Instance.Aws.EngineVersion)

		out, err := state.awsClient.CreateElastiCacheParameterGroup(ctx, name, family, []elasticachetypes.Tag{
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
			return awsmeta.LogErrorAndReturn(err, "Error creating parameter group", ctx)
		}

		logger = logger.WithValues("parameterGroupName", out.CacheParameterGroup.CacheParameterGroupName)
		logger.Info("Parameter group created")

		return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
	}
}

func createMainParameterGroup(state *State) composed.Action {
	return createParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
		GetAwsElastiCacheParameterGroupName(state.Obj().GetName()),
	)
}

func createTempParameterGroup(state *State) composed.Action {
	return createParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
		GetAwsElastiCacheTempParameterGroupName(state.Obj().GetName()),
	)
}
