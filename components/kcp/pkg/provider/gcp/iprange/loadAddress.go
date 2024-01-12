package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func loadAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP Address")

	//Get from GCP.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork
	list, err := state.computeClient.ListGlobalAddresses(ctx, project, vpc)
	if err != nil {
		return err, nil
	}

	//Iterate over the list and store the address in the state object
	for _, a := range list.Items {
		if ipRange.Name == a.Name {
			state.address = a
			break
		}
	}

	return nil, nil
}
