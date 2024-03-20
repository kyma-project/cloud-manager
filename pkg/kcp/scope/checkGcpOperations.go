package scope

import (
	"context"
	"fmt"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"os"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkGcpOperations(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	scope := state.ObjAsScope()
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
	var unfinishedOperations = make([]string, 0)
	for _, opName := range scope.Status.GcpOperations {
		logger.WithValues("Scope :", scope.Name).Info("Checking Service Enablement GCP Operation Status")
		op, err := client.GetServiceUsageOperation(ctx, opName)
		if err != nil {
			logger.Error(err, "Error getting Service Usage Operation from GCP. Retry via requeue.")
			return composed.StopWithRequeueDelay(gcpclient.GcpRetryWaitTime), nil
		}

		// This should not ever happen, but if it does, we can't do anything except removing the operation from the status
		if op == nil {
			logger.WithValues("Operation Name :", opName).Info("Operation not found in GCP.")
		}

		//Operation not completed yet
		if op != nil && !op.Done {
			unfinishedOperations = append(unfinishedOperations, opName)
		}

		if op != nil && op.Done {
			if op.Error != nil {
				logger.WithValues("Operation Name :", opName).Info("Operation failed. Removing from status.")
			}
		}

	}
	scope.Status.GcpOperations = unfinishedOperations
	if len(unfinishedOperations) > 0 {
		scope.Status.GcpOperations = unfinishedOperations
		return composed.UpdateStatus(scope).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	// Even if all operations are done (which might take us several iterations to figure out), we have issues with failed or not found operations.
	// We should requeue and verify that all APIs are really enabled.
	return composed.UpdateStatus(scope).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
