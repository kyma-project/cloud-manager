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

// utilizing subscriptionId as jobId
func (c restoreMockClient) TriggerRestore(_ context.Context,
	request RestoreRequest,
) (*string, error) {
	jobsMock.AddStorageJob(request.VaultName, request.TargetFolderName, armrecoveryservicesbackup.JobStatusInProgress)
	return nil, nil
}
