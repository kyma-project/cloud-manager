package nuke

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
)

func createAwsClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	roleName := fmt.Sprintf(
		// "arn:aws-us-gov:iam::%s:role/%s",
		"arn:aws:iam::%s:role/%s",
		state.Scope().Spec.Scope.Aws.AccountId,
		awsconfig.AwsConfig.Default.AssumeRoleName,
	)

	logger.
		WithValues(
			"awsRegion", state.Scope().Spec.Region,
			"awsRole", roleName,
		).
		Info("Assuming AWS role")

	cli, err := state.awsClientProvider(
		ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error assuming AWS role", err, ctx)
	}

	state.awsClient = cli

	return nil, nil
}
