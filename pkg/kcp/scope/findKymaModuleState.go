package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func findKymaModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	isListed := util.IsKymaModuleListedInSpec(state.kyma, "cloud-manager")

	if !isListed {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
