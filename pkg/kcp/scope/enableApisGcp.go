package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/serviceusage/v1"
	"os"
)

func enableApisGcp(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	saJsonKeyPath := os.Getenv("GCP_SA_JSON_KEY_PATH")
	if saJsonKeyPath == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("GCP_SA_JSON_KEY_PATH not set"),
			"Error enabling GCP APIs",
			composed.StopAndForget, // Without this env var, we can't call any APIs anyway, so we can't recover from this error
			ctx)
	}
	client, err := state.gcpServiceUsageClientProvider(ctx, saJsonKeyPath)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error getting ServiceUsageClient: %w", err),
			"Error enabling GCP APIs",
			composed.StopAndForget, // Without this env var, we can't call any APIs anyway, so we can't recover from this error
			ctx)
	}
	// Enable GCP APIs
	scope := state.ObjAsScope()
	// compute
	err, _ = verifyAndAddOperationToStatus(ctx, scope, client, gcpclient.ComputeService)
	if err != nil {
		return err, ctx
	}

	// service networking
	err, _ = verifyAndAddOperationToStatus(ctx, scope, client, gcpclient.ServiceNetworkingService)
	if err != nil {
		return err, ctx
	}

	// cloudresourcemanager
	err, _ = verifyAndAddOperationToStatus(ctx, scope, client, gcpclient.CloudResourceManagerService)
	if err != nil {
		return err, ctx
	}
	// filestore
	err, _ = verifyAndAddOperationToStatus(ctx, scope, client, gcpclient.FilestoreService)
	if err != nil {
		return err, ctx
	}

	if len(scope.Status.GcpOperations) == 0 {
		logger.Info("All APIs are enabled. Proceeding to next step.")
		return nil, nil
	}
	return composed.UpdateStatus(scope).
		SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpOperationWaitTime)).
		Run(ctx, state)
}

func verifyAndAddOperationToStatus(ctx context.Context, scope *v1beta1.Scope, client gcpclient.ServiceUsageClient, service gcpclient.GcpServiceName) (error, context.Context) {
	operation, err := verifyAndEnable(ctx, scope, client, service)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error enabling service %s: %w", gcpclient.ComputeService, err),
			"Error enabling GCP APIs",
			composed.StopAndForget,
			ctx)
	}
	scope.Status.GcpOperations = make([]string, 0)
	if operation != nil {
		scope.Status.GcpOperations = append(scope.Status.GcpOperations, operation.Name)
	}
	return nil, ctx
}

func verifyAndEnable(ctx context.Context, scope *v1beta1.Scope, client gcpclient.ServiceUsageClient, service gcpclient.GcpServiceName) (operation *serviceusage.Operation, err error) {
	logger := composed.LoggerFromCtx(ctx)
	enabled, err := client.IsServiceEnabled(ctx, scope.Spec.Scope.Gcp.Project, service)
	if err != nil {
		return nil, err
	}
	if !enabled {
		operation, err := client.EnableService(ctx, scope.Spec.Scope.Gcp.Project, service)
		if err != nil {
			return nil, err
		}
		logger.Info("Request to enable API submitted", "API", service, "Operation", operation.Name)
		return operation, nil
	}
	return nil, nil
}
