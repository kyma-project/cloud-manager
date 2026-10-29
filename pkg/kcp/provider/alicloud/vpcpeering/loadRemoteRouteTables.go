package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadRemoteRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	remoteNet := state.RemoteNetwork()
	if remoteNet.Status.Network == nil || remoteNet.Status.Network.Alicloud == nil {
		return nil, ctx
	}
	remoteVpcId := remoteNet.Status.Network.Alicloud.VpcId
	tables, err := state.remoteClient.DescribeRouteTables(ctx, remoteVpcId)
	if err != nil {
		// On delete the remote client may point at the wrong account/region (cross-account delete
		// reuses the local client), so a failed remote lookup must not wedge teardown — skip and continue.
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err, "Error loading remote route tables but skipping as marked for deletion", nil, ctx)
		}
		return composed.LogErrorAndReturn(err, "Error loading remote route tables", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.remoteRouteTables = tables

	return nil, ctx
}
