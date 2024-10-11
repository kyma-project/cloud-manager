package gcpvpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitRemoteNetworkCreation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.KcpRemoteNetwork.Status.Conditions) == 0 {
		return composed.StopWithRequeueDelay(3 * util.Timing.T1000ms()), nil
	}

	return nil, nil
}
