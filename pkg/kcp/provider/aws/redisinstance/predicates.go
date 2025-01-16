package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shouldCreateMainParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.parameterGroup == nil &&
			(state.elastiCacheReplicationGroup == nil ||
				state.elastiCacheReplicationGroup != nil && state.IsRedisVersionUpToDate())
	}
}

func shouldModifyMainParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.parameterGroup != nil && !state.AreMainParamGroupParamsUpToDate()
	}
}

func shouldCreateTempParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && state.tempParameterGroup == nil &&
			!state.IsMainParamGroupFamilyUpToDate()
	}
}

func shouldModifyTempParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && state.tempParameterGroup != nil &&
			!state.AreTempParamGroupParamsUpToDate()
	}
}

func shouldUpdateRedisPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)
		return state.ShouldUpdateRedisInstance()
	}
}

func shouldUpgradeRedisPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && !state.IsRedisVersionUpToDate() &&
			(state.tempParameterGroup != nil && state.AreTempParamGroupParamsUpToDate() || state.IsMainParamGroupFamilyUpToDate())
	}
}

func shouldDeleteObsoleteMainParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && state.tempParameterGroup != nil &&
			!state.IsMainParamGroupFamilyUpToDate() && state.AreTempParamGroupParamsUpToDate() && state.IsRedisVersionUpToDate() && !state.IsMainParamGroupUsed()
	}
}

func shouldSwitchToMainParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && state.tempParameterGroup != nil &&
			state.AreMainParamGroupParamsUpToDate() && state.IsMainParamGroupFamilyUpToDate() && state.IsRedisVersionUpToDate() && !state.IsMainParamGroupUsed()
	}
}

func shouldDeleteRedundantTempParamGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		return state.elastiCacheReplicationGroup != nil && state.parameterGroup != nil && state.tempParameterGroup != nil &&
			state.IsMainParamGroupFamilyUpToDate() && !state.IsTempParamGroupUsed()
	}
}
