package wafpolicy

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func detachLogging(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	webAcl := state.ObjAsWafPolicy()
	if webAcl.Status.ProviderId == "" {
		return nil, ctx
	}

	logger.Info("Detaching WebACL from log group", "webAclArn", webAcl.Status.ProviderId)

	err := state.awsClient.DeleteLoggingConfiguration(ctx, webAcl.Status.ProviderId)
	if err != nil {
		if awsmeta.IsNotFound(err) {
			logger.Info("No logging configuration found for WebACL, skipping detach")
			return nil, ctx
		}
		if awsmeta.IsErrorRetryable(err) {
			return composed.LogErrorAndReturn(err, "Error detaching WebACL logging (retryable)", composed.StopWithRequeue, ctx)
		}
		return composed.LogErrorAndReturn(err, "Error detaching WebACL logging", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL logging detached successfully")
	return nil, ctx
}
