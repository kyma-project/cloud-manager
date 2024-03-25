package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func predicateShouldEnable() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)

		// module is listed in Kyma status, but SKR is not active
		if state.moduleState != util.KymaModuleStateNotPresent && !state.skrActive {
			return true
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
