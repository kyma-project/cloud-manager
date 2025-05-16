package exposedData

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func addressesLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	for _, router := range state.routers {
		addresses, err := state.gcpClient.GetRouterIpAddresses(
			ctx,
			state.ObjAsScope().Spec.Scope.Gcp.Project,
			state.ObjAsScope().Spec.Region,
			ptr.Deref(router.Name, ""),
		)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading GCP addresses for the router in exposed data", composed.StopWithRequeue, ctx)
		}

		state.addresses = append(state.addresses, addresses...)
	}

	return nil, ctx
}
