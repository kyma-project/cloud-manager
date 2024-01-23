package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func saveScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	err := state.Cluster().K8sClient().Create(ctx, state.ObjAsScope())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating scope", composed.StopWithRequeue, nil)
	}

	return nil, nil
}
