package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func skrDeactivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.moduleState != util.KymaModuleStateNotPresent {
		return nil, nil
	}
	if state.ObjAsScope() == nil || state.ObjAsScope().GetName() == "" {
		return nil, nil
	}

	logger.Info("Stopping SKR")

	state.activeSkrCollection.RemoveScope(state.ObjAsScope())

	return nil, ctx
}
