package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func findKymaModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.moduleState = util.GetKymaModuleState(state.kyma, "cloud-manager")

	if state.moduleState != util.KymaModuleStateReady {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
