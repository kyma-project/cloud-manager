package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadRemoteVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	obj := state.ObjAsVpcPeering()

	logger := composed.LoggerFromCtx(ctx)

	// skip loading of vpc peering connections if remoteId is empty
	if len(obj.Status.RemoteId) == 0 {
		return nil, nil
	}

	remoteAccountId := obj.Spec.VpcPeering.Aws.RemoteAccountId
	remoteRegion := obj.Spec.VpcPeering.Aws.RemoteRegion

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger.WithValues(
		"remoteAwsRegion", remoteRegion,
		"remoteAwsRole", roleArn,
	).Info("Assuming remote AWS role")

	client, err := state.provider(
		ctx,
		remoteRegion,
		state.awsAccessKeyid,
		state.awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error initializing remote AWS client", ctx)
	}

	list, err := client.DescribeVpcPeeringConnections(ctx)

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing AWS peering connections", ctx)
	}

	for _, c := range list {
		if obj.Status.RemoteId == ptr.Deref(c.VpcPeeringConnectionId, "") {
			state.remoteVpcPeering = &c
			break
		}
	}

	ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("remoteId", obj.Status.RemoteId))

	if state.remoteVpcPeering == nil {
		return composed.LogErrorAndReturn(fmt.Errorf("error loading remote AWS VPC Peering"), "Error loading remote AWS VPC Peering", composed.StopAndForget, ctx)
	}

	return nil, ctx
}
