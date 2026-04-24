package awswebacl

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func ensureLogGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if logging disabled
	if webAcl.Spec.LoggingConfiguration == nil || !webAcl.Spec.LoggingConfiguration.Enabled {
		return nil, ctx
	}

	// Generate log group name using centralized method
	// Each WebACL gets its own dedicated log group
	logGroupName := ManagedLogGroupName(webAcl)
	state.managedLogGroupName = logGroupName

	scope := state.Scope()
	if scope == nil {
		return composed.LogErrorAndReturn(fmt.Errorf("scope not loaded"),
			"Cannot create log group", composed.StopWithRequeue, ctx)
	}

	logger.Info("Ensuring log group", "name", logGroupName)

	// Create (idempotent)
	if err := state.awsClient.CreateLogGroup(ctx, logGroupName); err != nil {
		return composed.LogErrorAndReturn(err, "Error creating log group", composed.StopWithRequeue, ctx)
	}

	// Set retention
	retentionDays := int32(30)
	if webAcl.Spec.LoggingConfiguration.Destination != nil &&
		webAcl.Spec.LoggingConfiguration.Destination.CloudWatchLogs != nil &&
		webAcl.Spec.LoggingConfiguration.Destination.CloudWatchLogs.RetentionDays > 0 {
		retentionDays = webAcl.Spec.LoggingConfiguration.Destination.CloudWatchLogs.RetentionDays
	}

	if err := state.awsClient.PutRetentionPolicy(ctx, logGroupName, retentionDays); err != nil {
		return composed.LogErrorAndReturn(err, "Error setting retention", composed.StopWithRequeue, ctx)
	}

	// Tag (best-effort)
	tags := map[string]string{
		"ManagedBy":  "cloud-manager",
		"ShootName":  scope.Spec.ShootName,
		"WebAclName": webAcl.Name,
	}
	_ = state.awsClient.TagLogGroup(ctx, logGroupName, tags)

	logger.Info("Log group ready", "retention", retentionDays)
	return nil, ctx
}
