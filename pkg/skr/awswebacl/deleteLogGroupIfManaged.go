package awswebacl

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteLogGroupIfManaged(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if not managed
	if webAcl.Status.LoggingStatus == nil || !webAcl.Status.LoggingStatus.ManagedLogGroup {
		return nil, ctx
	}

	// Each WebACL has its own dedicated log group
	logGroupName := ManagedLogGroupName(webAcl)
	logger.Info("Deleting managed log group", "name", logGroupName)

	if err := state.awsClient.DeleteLogGroup(ctx, logGroupName); err != nil {
		logger.Error(err, "Warning: failed to delete log group")
		// Don't fail deletion
	}

	return nil, ctx
}
