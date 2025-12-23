package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"io"
)

type RestoreClient interface {
	TriggerRestore(ctx context.Context,
		request RestoreRequest,
	) (*string, error)
}

type restoreClient struct {
	*armrecoveryservicesbackup.RestoresClient
}

func NewRestoreClient(rc *armrecoveryservicesbackup.RestoresClient) RestoreClient {
	return restoreClient{rc}
}

type TriggerResponse struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type RestoreRequest struct {
	VaultName                string
	ResourceGroupName        string
	FabricName               string
	ContainerName            string
	ProtectedItemName        string
	RecoveryPointId          string
	SourceStorageAccountPath string
	TargetStorageAccountPath string
	TargetFileShareName      string
	TargetFolderName         string
}

func (c restoreClient) TriggerRestore(ctx context.Context,
	request RestoreRequest,
) (*string, error) {
	logger := composed.LoggerFromCtx(ctx).WithName("restoreClient - TriggerRestore")
	parameters := armrecoveryservicesbackup.RestoreRequestResource{
		Properties: to.Ptr(armrecoveryservicesbackup.AzureFileShareRestoreRequest{
			ObjectType:   to.Ptr("AzureFileShareRestoreRequest"),
			CopyOptions:  to.Ptr(armrecoveryservicesbackup.CopyOptionsOverwrite),
			RecoveryType: to.Ptr(armrecoveryservicesbackup.RecoveryTypeAlternateLocation),
			RestoreFileSpecs: []*armrecoveryservicesbackup.RestoreFileSpecs{
				{
					TargetFolderPath: to.Ptr(request.TargetFolderName),
				},
			},
			RestoreRequestType: to.Ptr(armrecoveryservicesbackup.RestoreRequestTypeFullShareRestore),
			SourceResourceID:   to.Ptr(request.SourceStorageAccountPath), // Source file share arm id
			TargetDetails: to.Ptr(armrecoveryservicesbackup.TargetAFSRestoreInfo{
				Name:             to.Ptr(request.TargetFileShareName),      // Target File share name
				TargetResourceID: to.Ptr(request.TargetStorageAccountPath), // Target file share arm id
			}),
		}),
	}

	poller, err := c.BeginTrigger(
		ctx,
		request.VaultName,
		request.ResourceGroupName,
		request.FabricName,
		request.ContainerName,
		request.ProtectedItemName,
		request.RecoveryPointId,
		parameters,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger the restore operation: %w", err)
	}

	// Poll the restore operation
	resp, err := poller.Poll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to poll the restore operation: %w", err)
	}

	// Read resp body
	if resp == nil || resp.Body == nil {
		return nil, errors.New("response body is nil")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}
	var data TriggerResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the response body: %w", err)
	}
	logger.Info("TriggerRestore response", "jobId", data.Id)
	return &data.Id, nil
}
