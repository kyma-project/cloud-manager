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
		subscriptionId string,
		vaultName string,
		resourceGroupName string,
		fabricName string,
		containerName string,
		protectedItemName string,
		recoveryPointId string,
		location string,
		sourceStorageAccountName string,
		targetStorageAccountName string,
		targetFileShareName string,
		targetFolderName string,
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

func (c restoreClient) TriggerRestore(ctx context.Context,
	subscriptionId string,
	vaultName string,
	resourceGroupName string,
	fabricName string,
	containerName string,
	protectedItemName string,
	recoveryPointId string,
	location string,
	sourceStorageAccountName string,
	targetStorageAccountName string,
	targetFileShareName string,
	targetFolderName string,
) (*string, error) {
	sourceResourceId := GetStorageAccountPath(subscriptionId, resourceGroupName, sourceStorageAccountName)
	targetResourceId := GetStorageAccountPath(subscriptionId, resourceGroupName, targetStorageAccountName)
	parameters := armrecoveryservicesbackup.RestoreRequestResource{
		ETag:     nil,
		Location: to.Ptr(location),
		Properties: to.Ptr(armrecoveryservicesbackup.AzureFileShareRestoreRequest{
			ObjectType:                     to.Ptr("AzureFileShareRestoreRequest"),
			CopyOptions:                    to.Ptr(armrecoveryservicesbackup.CopyOptionsOverwrite),
			RecoveryType:                   to.Ptr(armrecoveryservicesbackup.RecoveryTypeAlternateLocation),
			ResourceGuardOperationRequests: nil,
			RestoreFileSpecs: []*armrecoveryservicesbackup.RestoreFileSpecs{
				{
					TargetFolderPath: to.Ptr(targetFolderName),
				},
			},
			RestoreRequestType: to.Ptr(armrecoveryservicesbackup.RestoreRequestTypeFullShareRestore),
			SourceResourceID:   to.Ptr(sourceResourceId), // not sure if correct
			TargetDetails: to.Ptr(armrecoveryservicesbackup.TargetAFSRestoreInfo{
				Name:             to.Ptr(targetFileShareName), // Target File share name
				TargetResourceID: to.Ptr(targetResourceId),    // Target file share arm id; try nil for now
			}),
		}),
		Tags: nil,
		ID:   nil,
		Name: nil,
		Type: nil,
	}

	poller, err := c.BeginTrigger(
		ctx,
		vaultName,
		resourceGroupName,
		fabricName,
		containerName,
		protectedItemName,
		recoveryPointId,
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
