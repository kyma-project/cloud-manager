package awswebacl

import (
	"context"
	"errors"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadLoggingConfiguration(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	loggingConfig, err := state.awsClient.GetLoggingConfiguration(ctx, webAcl.Status.Arn)
	if err != nil {
		var notFound *wafv2types.WAFNonexistentItemException
		if errors.As(err, &notFound) {
			state.loggingConfig = nil
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading logging configuration", composed.StopWithRequeue, ctx)
	}

	state.loggingConfig = loggingConfig
	logger.Info("Logging configuration loaded")
	return nil, ctx
}
