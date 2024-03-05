package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func findKymaModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// Once module is added to the SKR Kyma CR, in KCP it first appears in the status field
	// with state: Processing, and it does not appear in the spec
	moduleState := util.GetKymaModuleStateFromStatus(state.kyma, "cloud-manager")
	//isListed := util.IsKymaModuleListedInSpec(state.kyma, "cloud-manager")
	isListed := moduleState != util.KymaModuleStateNotPresent

	if !isListed {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
