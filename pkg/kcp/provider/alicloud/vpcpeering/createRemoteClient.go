package vpcpeering

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createRemoteClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	remoteNet := state.RemoteNetwork()
	if remoteNet.Status.Network == nil || remoteNet.Status.Network.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("remote network Alicloud status not yet populated"),
			"Waiting for remote network status before creating remote client",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	remoteAccountId := remoteNet.Status.Network.Alicloud.AccountId
	remoteRegion := remoteNet.Status.Network.Alicloud.Region

	state.remoteAccountId = remoteAccountId
	state.remoteRegion = remoteRegion

	if remoteAccountId == state.localAccountId {
		state.remoteClient = state.client
		return nil, ctx
	}

	// Cross-account peering requires assumeRoleArn support (pending rebase of #2199 into this branch).
	return composed.LogErrorAndReturn(
		fmt.Errorf("cross-account AliCloud VPC peering not yet supported: remote account %s differs from local account %s", remoteAccountId, state.localAccountId),
		"Cross-account AliCloud VPC peering requires assumeRoleArn support (pending)",
		composed.StopWithRequeueDelay(util.Timing.T300000ms()),
		ctx,
	)
}
