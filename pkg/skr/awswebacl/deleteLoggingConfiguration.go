package awswebacl

import (
	"context"
	"errors"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteLoggingConfiguration(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	logger.Info("Deleting logging configuration")
	err := state.awsClient.DeleteLoggingConfiguration(ctx, webAcl.Status.Arn)
	if err != nil {
		var notFound *wafv2types.WAFNonexistentItemException
		if errors.As(err, &notFound) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error deleting logging", composed.StopWithRequeue, ctx)
	}

	logger.Info("Logging deleted")
	return nil, ctx
}
