package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		redisInstanceState := st.(types.State)
		state, err := stateFactory.NewState(ctx, redisInstanceState)
		if err != nil {
			err = fmt.Errorf("error creating new aws rediscluster state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"awsRedisCluster",
			actions.AddCommonFinalizer(),
			loadSubnetGroup,
			loadMainParameterGroup(state),
			loadTempParameterGroup(state),
			loadMainParameterGroupCurrentParams(),
			loadTempParameterGroupCurrentParams(),
			loadMainParameterGroupFamilyDefaultParams(),
			loadTempParameterGroupFamilyDefaultParams(),
			loadAuthTokenSecret,
			loadUserGroup,
			findSecurityGroup,
			loadSecurityGroup,
			loadElastiCacheCluster,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"redisInstance-create",
					loadMemberClusters,
					createSubnetGroup,
					composed.If(
						shouldDeleteObsoleteMainParamGroupPredicate(),
						deleteMainParameterGroup(),
					),
					composed.If(
						shouldDeleteRedundantTempParamGroupPredicate(),
						deleteTempParameterGroup(),
					),
					composed.If(
						shouldCreateMainParamGroupPredicate(),
						createMainParameterGroup(state),
					),
					composed.If(
						shouldCreateTempParamGroupPredicate(),
						createTempParameterGroup(state),
					),
					composed.If(
						shouldModifyMainParamGroupPredicate(),
						modifyMainParameterGroup(state),
					),
					composed.If(
						shouldModifyTempParamGroupPredicate(),
						modifyTempParameterGroup(state),
					),
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
					composed.If(
						shouldUpdateRedisPredicate(),
						updateElastiCacheCluster(),
					),
					composed.If(
						shouldUpgradeRedisPredicate(),
						upgradeElastiCacheCluster(),
					),
					composed.If(
						shouldSwitchToMainParamGroupPredicate(),
						switchToMainParamGroup(),
					),
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
					deleteMainParameterGroup(),
					deleteTempParameterGroup(),
					deleteSubnetGroup,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(awsmeta.SetAwsAccountId(ctx, redisInstanceState.Scope().Spec.Scope.Aws.AccountId), state)
	}
}
