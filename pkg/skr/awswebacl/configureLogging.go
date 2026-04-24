package awswebacl

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func configureLogging(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	// Scenario 1: Logging disabled - delete if exists
	if webAcl.Spec.LoggingConfiguration == nil || !webAcl.Spec.LoggingConfiguration.Enabled {
		if state.loggingConfig != nil {
			logger.Info("Disabling logging")
			if err := state.awsClient.DeleteLoggingConfiguration(ctx, webAcl.Status.Arn); err != nil {
				return composed.LogErrorAndReturn(err, "Error disabling logging", composed.StopWithRequeue, ctx)
			}
			state.loggingConfig = nil

			// Delete managed log group if it was managed by this WebACL
			if webAcl.Status.LoggingStatus != nil && webAcl.Status.LoggingStatus.ManagedLogGroup {
				logGroupName := ManagedLogGroupName(webAcl)
				logger.Info("Deleting managed log group", "name", logGroupName)
				if err := state.awsClient.DeleteLogGroup(ctx, logGroupName); err != nil {
					logger.Error(err, "Warning: failed to delete log group")
					// Don't fail reconciliation, log group cleanup is best-effort
				}
			}
		}
		return nil, ctx
	}

	// Scenario 2: Build and apply logging config
	desiredConfig, err := buildLoggingConfiguration(webAcl, state)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error building logging config", composed.StopWithRequeue, ctx)
	}

	// Check if update needed
	if state.loggingConfig != nil && loggingConfigMatches(state.loggingConfig, desiredConfig) {
		logger.Info("Logging up to date")
		return nil, ctx
	}

	// Put (create or update)
	logger.Info("Configuring logging")
	if err := state.awsClient.PutLoggingConfiguration(ctx, &wafv2.PutLoggingConfigurationInput{
		LoggingConfiguration: desiredConfig,
	}); err != nil {
		return composed.LogErrorAndReturn(err, "Error configuring logging", composed.StopWithRequeue, ctx)
	}

	logger.Info("Logging configured", "destination", desiredConfig.LogDestinationConfigs)
	return nil, ctx
}

func buildLoggingConfiguration(webAcl *cloudresourcesv1beta1.AwsWebAcl, state *State) (*wafv2types.LoggingConfiguration, error) {
	config := &wafv2types.LoggingConfiguration{
		ResourceArn: ptr.To(webAcl.Status.Arn),
	}

	// Build log destination ARN for auto-managed log group
	if state.managedLogGroupName == "" {
		return nil, fmt.Errorf("managed log group name not set")
	}
	scope := state.Scope()
	if scope == nil {
		return nil, fmt.Errorf("scope not loaded")
	}
	accountId := getAwsAccountId(scope)
	logDestinationArn := fmt.Sprintf(
		"arn:aws:logs:%s:%s:log-group:%s",
		scope.Spec.Region,
		accountId,
		state.managedLogGroupName,
	)

	config.LogDestinationConfigs = []string{logDestinationArn}

	// Redacted fields
	if len(webAcl.Spec.LoggingConfiguration.RedactedFields) > 0 {
		redacted, err := convertRedactedFields(webAcl.Spec.LoggingConfiguration.RedactedFields)
		if err != nil {
			return nil, err
		}
		config.RedactedFields = redacted
	}

	return config, nil
}

func loggingConfigMatches(current, desired *wafv2types.LoggingConfiguration) bool {
	if len(current.LogDestinationConfigs) != len(desired.LogDestinationConfigs) {
		return false
	}
	if len(current.LogDestinationConfigs) > 0 &&
		current.LogDestinationConfigs[0] != desired.LogDestinationConfigs[0] {
		return false
	}
	return reflect.DeepEqual(current.RedactedFields, desired.RedactedFields)
}

func getAwsAccountId(scope *cloudcontrolv1beta1.Scope) string {
	if scope.Spec.Scope.Aws != nil {
		return scope.Spec.Scope.Aws.AccountId
	}
	return ""
}
