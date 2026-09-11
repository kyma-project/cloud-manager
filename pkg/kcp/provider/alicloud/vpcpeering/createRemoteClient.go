package vpcpeering

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
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

	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret

	c, err := state.provider(ctx, remoteRegion, accessKeyId, accessKeySecret)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error creating alicloud remote vpcpeering client: %w", err),
			"Error creating remote client",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}
	state.remoteClient = c

	return nil, ctx
}
