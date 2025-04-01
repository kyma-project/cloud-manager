package kyma

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func kymaFindModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.skrActive = state.activeSkrCollection.Contains(state.Name().Name)

	isKymaLoaded := false

	if state.ObjAsKyma() == nil {
		state.moduleState = util.KymaModuleStateNotPresent
		state.moduleInSpec = false
	} else {
		isKymaLoaded = true
		moduleName := "cloud-manager"
		state.moduleState = util.GetKymaModuleStateFromStatus(state.ObjAsKyma(), moduleName)
		state.moduleInSpec = util.IsKymaModuleListedInSpec(state.ObjAsKyma(), moduleName)
	}

	logger := composed.LoggerFromCtx(ctx)
	logger = logger.WithValues(
		"kymaLoaded", fmt.Sprintf("%v", isKymaLoaded),
		"moduleState", state.moduleState,
		"moduleInSpec", fmt.Sprintf("%v", state.moduleInSpec),
		"skrActive", fmt.Sprintf("%v", state.skrActive),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
