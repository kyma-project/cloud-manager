package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
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

	peering, err := state.client.DescribeVpcPeeringConnection(ctx, obj.Status.Id)

	if awsmeta.IsNotFound(err) {
		logger.Info("Local AWS Peering not found for KCP VpcPeering")
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS peering connections", composed.StopWithRequeue, ctx)
	}

	state.vpcPeering = peering

	if state.vpcPeering == nil {
		return nil, nil
	}

	logger = logger.WithValues(
		"id", ptr.Deref(state.vpcPeering.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
