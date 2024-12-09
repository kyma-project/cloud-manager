package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadRemoteVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	// skip loading of vpc peering connections if remoteId is empty
	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	peering, err := state.remoteClient.DescribeVpcPeeringConnection(ctx, obj.Status.RemoteId)

	if err != nil {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err,
				"Error listing AWS peering connections but skipping as marked for deletion",
				nil,
				ctx)
		}
		return awsmeta.LogErrorAndReturn(err, "Error listing AWS peering connections", ctx)
	}

	state.remoteVpcPeering = peering

	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("remoteId", obj.Status.RemoteId))

	if state.remoteVpcPeering == nil {
		return composed.LogErrorAndReturn(fmt.Errorf("error loading remote AWS VPC Peering"), "Error loading remote AWS VPC Peering", composed.StopAndForget, ctx)
	}

	return nil, ctx
}
