package exposedData

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func routersLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	routers, err := state.gcpClient.GetVpcRouters(
		ctx,
		state.ObjAsScope().Spec.Scope.Gcp.Project,
		state.ObjAsScope().Spec.Region,
		state.ObjAsScope().Spec.Scope.Gcp.VpcNetwork,
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading GCP Routers for exposedData", composed.StopWithRequeue, ctx)
	}

	state.routers = routers

	return nil, ctx
}
