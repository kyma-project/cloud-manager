package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"time"
)

func createAwsClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	roleName := fmt.Sprintf(
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
		state.Scope().Spec.Region,
		state.env.Get("AWS_ACCESS_KEY_ID"),
		state.env.Get("AWS_SECRET_ACCESS_KEY"),
		roleName,
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error assuming AWS role", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	state.awsClient = cli

	return nil, nil
}
