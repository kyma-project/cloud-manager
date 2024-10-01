package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func findKymaModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kyma == nil {
		state.moduleState = util.KymaModuleStateNotPresent
		return nil, ctx
	}

	moduleName := "cloud-manager"
	moduleState := util.GetKymaModuleStateFromStatus(state.kyma, moduleName)
	moduleInSpec := util.IsKymaModuleListedInSpec(state.kyma, moduleName)
	skrActive := state.activeSkrCollection.Contains(state.kyma.GetName())

	logger = logger.WithValues(
		"moduleState", moduleState,
		"moduleInSpec", fmt.Sprintf("%v", moduleInSpec),
		"skrActive", fmt.Sprintf("%v", skrActive),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.moduleState = moduleState
	state.moduleInSpec = moduleInSpec
	state.skrActive = skrActive

	return nil, ctx
}
