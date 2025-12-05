package scope

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/googleapi"
)

func checkGcpOperations(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	scope := state.ObjAsScope()

	if config.GcpConfig.CredentialsFile == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("gcpConfig.credentialsFile not set"),
			"Error enabling GCP APIs",
			composed.StopAndForget, // Without this env var, we can't call any APIs anyway, so we can't recover from this error
			ctx)
	}
	client, err := state.gcpServiceUsageClientProvider(ctx, config.GcpConfig.CredentialsFile)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error getting ServiceUsageClient: %w", err),
			"Error enabling GCP APIs",
			composed.StopAndForget, // Without this env var, we can't call any APIs anyway, so we can't recover from this error
			ctx)
	}
	var unfinishedOperations = make([]string, 0)
	for _, opName := range scope.Status.GcpOperations {
		logger.WithValues("Scope", scope.Name).Info("Checking Service Enablement GCP Operation Status")
		op, err := client.GetServiceUsageOperation(ctx, opName)
		//If the operation is not found, reset the operation to retry
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				logger.WithValues("operationName", opName).Info("Operation not found in GCP. "+
					"Removing operation id from status and retry.", "error", err)
				continue
				// By not adding it to unfinished, it will be removed and retried
			}
			if e.Code == 400 {
				logger.WithValues("operationName", opName).
					Info("Operation failed because of invalid request. "+
						"This can happen because of invalid operation id. "+
						"Removing operation id from status and retry.", "error", err)
				// By not adding it to unfinished, it will be removed and retried
				continue

			}
		}
		if err != nil {
			logger.Error(err, "Error getting Service Usage Operation from GCP. Retry via requeue.")
			return composed.StopWithRequeueDelay(gcpclient.GcpRetryWaitTime), nil
		}

		// This should not ever happen, but if it does, we can't do anything except removing the operation from the status
		if op == nil {
			logger.WithValues("operationName", opName).Info("Operation not found in GCP.")
		}

		//Operation not completed yet
		if op != nil && !op.Done {
			unfinishedOperations = append(unfinishedOperations, opName)
		}

		if op != nil && op.Done {
			if op.Error != nil {
				logger.WithValues("operationName", opName).
					Info("Operation failed. Removing from status to retry.", "error message",
						op.Error.Message, "error code", op.Error.Code)
			}
		}

	}
	scope.Status.GcpOperations = unfinishedOperations
	if len(unfinishedOperations) > 0 {
		scope.Status.GcpOperations = unfinishedOperations
		return composed.PatchStatus(scope).
			ErrorLogMessage("Error patching KCP Scope status with GCP unfinished operations").
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	// Even if all operations are done (which might take us several iterations to figure out), we have issues with failed or not found operations.
	// We should requeue and verify that all APIs are really enabled.
	return composed.PatchStatus(scope).
		ErrorLogMessage("Error patching KCP Scope status with GCP operations").
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T100ms())).
		Run(ctx, state)
}
