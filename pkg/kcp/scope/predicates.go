package scope

import (
	"context"
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
			if !ObjIsLoadedPredicate()(ctx, state) {
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
