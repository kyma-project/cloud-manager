package exposedData

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func routerLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	routerName := state.ObjAsScope().Spec.Scope.OpenStack.VpcNetwork

	router, err := state.sapClient.GetRouterByName(ctx, routerName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting SAP router", composed.StopWithRequeue, ctx)
	}

	state.router = router

	return nil, ctx
}
