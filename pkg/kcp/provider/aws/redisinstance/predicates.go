package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
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

// shouldCreateTransientUserGroupPredicate fires exactly when the reconciler
// needs to attach a transient RBAC user group to disable the cluster's AUTH
// token (i.e. spec.authEnabled just flipped true → false). The user group is
// a shim: AWS ElastiCache requires a user group to be attached in the same
// ModifyReplicationGroup call that deletes the auth token, but nothing
// long-lived needs it. Once the delete-side predicate detects the group has
// been detached, deleteUserGroup + waitUserGroupDeleted drives the account
// back to zero cm-<name> user groups.
//
// Returns true iff (a) the replication group has been loaded, (b) the current
// AWS-side auth is enabled, (c) the desired spec auth is disabled, and
// (d) no user group has been observed yet. Otherwise false — either there is
// nothing to downgrade, or the shim already exists.
func shouldCreateTransientUserGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		if state.elastiCacheReplicationGroup == nil {
			return false
		}
		if state.userGroup != nil {
			return false
		}
		currentAuth := ptr.Deref(state.elastiCacheReplicationGroup.AuthTokenEnabled, false)
		desiredAuth := state.ObjAsRedisInstance().Spec.Instance.Aws.AuthEnabled

		return currentAuth && !desiredAuth
	}
}

// isUserGroupAttached reports whether the observed replication group currently
// lists our transient cm-<name> user group among its UserGroupIds. Used by
// shouldDeleteTransientUserGroupPredicate to distinguish "attached, wait for
// modify to detach" from "detached, safe to delete".
func isUserGroupAttached(state *State) bool {
	if state.userGroup == nil || state.elastiCacheReplicationGroup == nil {
		return false
	}
	name := ptr.Deref(state.userGroup.UserGroupId, "")
	for _, id := range state.elastiCacheReplicationGroup.UserGroupIds {
		if id == name {
			return true
		}
	}
	return false
}

// shouldDeleteTransientUserGroupPredicate fires when the transient user group
// exists AND has already been detached from the replication group. This
// invariant covers every relevant case:
//
//   - Downgrade in progress: mid-flow the UG is attached, so predicate returns
//     false; only after the Remove modify propagates does it fire.
//   - Upgrade with a stale leftover UG: predicate fires and cleans up. The
//     ROTATE modify does not re-attach the UG, so a detached UG present at
//     upgrade start is safe to delete.
//   - Steady-state orphan on an auth-enabled cluster: predicate fires
//     (inline backfill of legacy detached-orphan user groups).
func shouldDeleteTransientUserGroupPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		if state.userGroup == nil {
			return false
		}
		return !isUserGroupAttached(state)
	}
}
