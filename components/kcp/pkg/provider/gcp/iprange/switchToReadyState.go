package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func switchToReadyState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If desiredState != actualState, continue.
	if !state.inSync {
		return nil, nil
	}

	//If desiredState == actualState, update the status, and stop reconcilation
	err := state.AddReadyCondition(ctx, "IPRange provisioned in GCP")
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating IpRange success status", composed.StopWithRequeue, nil)
	}

	return composed.StopAndForget, nil
}
