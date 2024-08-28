package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadRemoteRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	remoteAccountId := obj.Spec.VpcPeering.Aws.RemoteAccountId
	remoteRegion := obj.Spec.VpcPeering.Aws.RemoteRegion
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", remoteAccountId, state.roleName)

	logger.WithValues(
		"remoteAwsRegion", remoteRegion,
		"remoteAwsRole", roleArn,
	).Info("Assuming remote AWS role")

	ctx = composed.LoggerIntoCtx(ctx, logger)

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

	routeTables, err := client.DescribeRouteTables(ctx, *state.remoteVpc.VpcId)

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS remote route tables", ctx)
	}

	state.remoteRouteTables = routeTables

	return nil, nil
}
