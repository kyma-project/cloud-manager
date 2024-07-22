package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func loadVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	// skip loading of vpc peering connections if connectionId is empty
	if len(obj.Status.Id) == 0 {
		return nil, nil
	}

	list, err := state.client.DescribeVpcPeeringConnections(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS peering connections", composed.StopWithRequeue, ctx)
	}

	for _, c := range list {
		if obj.Status.Id == ptr.Deref(c.VpcPeeringConnectionId, "") {
			state.vpcPeering = &c
			break
		}
	}

	if state.vpcPeering == nil {
		return nil, nil
	}

	logger = logger.WithValues(
		"vpcConnectionId", ptr.Deref(state.vpcPeering.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
