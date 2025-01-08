package redisinstance

import (
	"context"
	"fmt"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		redisInstanceState := st.(types.State)
		state, err := stateFactory.NewState(ctx, redisInstanceState)
		if err != nil {
			err = fmt.Errorf("error creating new aws redisinstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"awsRedisInstance",
			actions.AddFinalizer,
			loadSubnetGroup,
			loadParameterGroup,
			loadAuthTokenSecret,
			loadUserGroup,
			findSecurityGroup,
			loadSecurityGroup,
			loadElastiCacheCluster,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"redisInstance-create",
					createSubnetGroup,
					createParameterGroup(func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup }, GetAwsElastiCacheParameterGroupName(state.Obj().GetName())),
					modifyParameterGroup(func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup }, GetAwsElastiCacheParameterGroupName(state.Obj().GetName())),
					createAuthTokenSecret,
					createUserGroup,
					createSecurityGroup,
					authorizeSecurityGroupIngress,
					createElastiCacheCluster,
					updateStatusId,
					addUpdatingCondition,
					waitElastiCacheAvailable,
					waitUserGroupActive,
					modifyCacheNodeType,
					modifyAutoMinorVersionUpgrade,
					modifyPreferredMaintenanceWindow,
					modifyAuthEnabled,
					updateElastiCacheCluster,
					updateStatus,
				),
				composed.ComposeActions(
					"redisInstance-delete",
					removeReadyCondition,
					deleteElastiCacheCluster,
					waitElastiCacheDeleted,
					deleteSecurityGroup,
					deleteUserGroup,
					waitUserGroupDeleted,
					deleteAuthTokenSecret,
					deleteParameterGroup(func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup }, GetAwsElastiCacheParameterGroupName(state.Obj().GetName())),
					deleteSubnetGroup,
					actions.RemoveFinalizer,
				),
			),
			composed.StopAndForgetAction,
		)(awsmeta.SetAwsAccountId(ctx, redisInstanceState.Scope().Spec.Scope.Aws.AccountId), state)
	}
}
