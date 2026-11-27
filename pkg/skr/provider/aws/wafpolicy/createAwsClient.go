package wafpolicy

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
)

func createAwsClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	roleName := awsutil.RoleArnDefault(state.Scope().Spec.Scope.Aws.AccountId)

	cli, err := state.awsClientProvider(
		ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error assuming AWS role", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	state.awsClient = cli

	return nil, ctx
}
