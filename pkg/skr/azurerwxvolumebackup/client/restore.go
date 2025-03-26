package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"io"
	"log"
)

type RestoreClient interface {
	TriggerRestore(ctx context.Context,
		request RestoreRequest,
	) (*string, error)
}

type restoreClient struct {
	*armrecoveryservicesbackup.RestoresClient
}

func NewRestoreClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (RestoreClient, error) {

	rc, err := armrecoveryservicesbackup.NewRestoresClient(subscriptionId, cred, nil)

	if err != nil {
		return nil, err
	}

	return restoreClient{rc}, nil
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
		log.Println("failed to trigger the restore operation: " + err.Error())
		return nil, err
	}
	if poller.Done() {
		log.Println("poller is done")
	}
	log.Println("poller is not done")
	resp, err := poller.Poll(ctx)
	if err != nil {
		log.Println("failed to poll the restore operation: " + err.Error())
		return nil, err
	}
	if resp != nil {
		log.Println(resp.Status)
	}
	// Read resp body
	if resp == nil || resp.Body == nil {
		return nil, errors.New("response body is nil")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read the response body: " + err.Error())
		return nil, err
	}
	var data TriggerResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("failed to unmarshal the response body: " + err.Error())
		return nil, err
	}
	log.Println("jobId" + data.Id)
	return &data.Id, nil
}
