package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func predicateShouldEnable() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		// module is listed in Kyma status
		if state.moduleState != util.KymaModuleStateNotPresent {
			// but SKR is not active
			if !state.skrActive {
				return true
			}
			// or scope does not exist
			if !predicateScopeExists()(ctx, state) {
				return true
			}
			// or kyma network does not exist
			if state.kcpNetworkKyma == nil {
				return true
			}
		}

		return false
	}
}

func predicateShouldDisable() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		if state.moduleState == util.KymaModuleStateNotPresent && state.skrActive {
			return true
		}

		return false
	}
}

func predicateScopeExists() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)
		scopeObj := state.ObjAsScope()
		return scopeObj != nil && scopeObj.GetName() != "" // empty object is created when state gets created
	}
}

func predicateScopeCreateOrUpdateNeeded() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		if predicateScopeExists()(ctx, st) {
			state := st.(*State)

			// check if kyma network reference is created
			if state.kcpNetworkKyma == nil {
				return true
			}

			// check if labels from Kyma are copied to Scope
			for _, label := range cloudcontrolv1beta1.ScopeLabels {
				if _, ok := state.ObjAsScope().Labels[label]; !ok {
					return true
				}
			}

			// check provider specific shapes
			switch state.ObjAsScope().Spec.Provider {
			case cloudcontrolv1beta1.ProviderGCP:
				return len(state.ObjAsScope().Spec.Scope.Gcp.Workers) == 0
			case cloudcontrolv1beta1.ProviderAzure:
				return state.ObjAsScope().Spec.Scope.Azure.Network.Nodes == ""
			default:
				return false
			}
		} else {
			return true
		}
	}
}
