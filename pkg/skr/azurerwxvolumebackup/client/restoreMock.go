package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

var (
	restoreMock = &restoreMockClient{
		restoreClient: *newRestoreMockClient(),
	}
)

func newRestoreMockClient() *restoreClient {
	return &restoreClient{}
}

type restoreMockClient struct {
	restoreClient
}

func (c restoreMockClient) TriggerRestore(_ context.Context,
	request RestoreRequest,
) (*string, error) {
	// TriggerRestore utilizes vaultName as jobId and resourceGroupName as the final status after InProgress

	nextJobStatus := armrecoveryservicesbackup.JobStatusCompleted
	if request.ResourceGroupName != "" {
		nextJobStatus = armrecoveryservicesbackup.JobStatus(request.ResourceGroupName)
	}
	jobsMock.AddRestoreJob(request.VaultName, request.TargetFolderName, armrecoveryservicesbackup.JobStatusInProgress, nextJobStatus)
	return &request.VaultName, nil
}
