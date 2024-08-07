package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func loadRemoteVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	obj := state.ObjAsVpcPeering()

	// skip loading of vpc peering connections if remoteId is empty
	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	remoteAccountId := obj.Spec.VpcPeering.Aws.RemoteAccountId
	remoteRegion := obj.Spec.VpcPeering.Aws.RemoteRegion

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	client, err := state.provider(
		ctx,
		remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error initializing remote AWS client", composed.StopWithRequeue, ctx)
	}

	list, err := client.DescribeVpcPeeringConnections(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS peering connections", composed.StopWithRequeue, ctx)
	}

	for _, c := range list {
		if obj.Status.RemoteId == ptr.Deref(c.VpcPeeringConnectionId, "") {
			state.remoteVpcPeering = &c
			break
		}
	}

	logger := composed.LoggerFromCtx(ctx)
	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("remoteId", obj.Status.RemoteId))

	if state.remoteVpcPeering == nil {
		return composed.LogErrorAndReturn(fmt.Errorf("error loading remote AWS VPC Peering"), "Error loading remote AWS VPC Peering", composed.StopAndForget, ctx)
	}

	return nil, ctx
}
