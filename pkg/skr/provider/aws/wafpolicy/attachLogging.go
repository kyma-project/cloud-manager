package wafpolicy

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
)

func attachLogging(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.awsWebAcl == nil || state.awsWebAcl.ARN == nil {
		return nil, ctx
	}

	webAclArn := *state.awsWebAcl.ARN
	logGroupArn := awsutil.WafLogGroupArn(
		state.Scope().Spec.Region,
		state.Scope().Spec.Scope.Aws.AccountId,
	)

	if state.loggingConfig != nil {
		return nil, ctx
	}

	logger.Info("Attaching WebACL to log group", "webAclArn", webAclArn, "logGroupArn", logGroupArn)

	err := state.awsClient.PutLoggingConfiguration(ctx, webAclArn, logGroupArn)
	if err != nil {
		if awsmeta.IsErrorRetryable(err) {
			return composed.LogErrorAndReturn(err, "Error attaching WebACL logging (retryable)", composed.StopWithRequeue, ctx)
		}
		return composed.LogErrorAndReturn(err, "Error attaching WebACL logging", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL logging attached successfully")
	return nil, ctx
}
