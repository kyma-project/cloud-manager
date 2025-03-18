package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"strings"
)

type JobsClient interface {
	FindRestoreJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startFilter string, restoreFolderPath string) (*string, bool, error)
	GetStorageJob(ctx context.Context, vaultName string, resourceGroupName string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error)
}
type jobsClient struct {
	*armrecoveryservicesbackup.BackupJobsClient
	*armrecoveryservicesbackup.JobDetailsClient
}

func NewJobsClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (JobsClient, error) {
	bjc, err := armrecoveryservicesbackup.NewBackupJobsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	jdc, err := armrecoveryservicesbackup.NewJobDetailsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	return jobsClient{bjc, jdc}, nil
}

// FindRestoreJobId finds the restore job for the given vault, resource group, start filter and restore folder path
// If a job is in progress and the destination folder is not populated yet, it returns true as second return value which indicates that this should be retried at a later time
func (c jobsClient) FindRestoreJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startFilter string, restoreFolderPath string) (*string, bool, error) {
	backupJobsClientOptions := armrecoveryservicesbackup.BackupJobsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("backupManagementType eq 'AzureStorage' and operation eq 'Restore' and startTime ge %v", startFilter)),
	}
	pager := c.BackupJobsClient.NewListPager(vaultName, resourceGroupName, &backupJobsClientOptions)
	retry := false
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("error getting next page: %w", err)
		}
		for _, job := range page.Value {
			statusStr := job.Properties.GetJob().Status
			jobName := *job.Name
			entityFriendlyName := *job.Properties.GetJob().EntityFriendlyName
			if entityFriendlyName != fileShareName {
				continue
			}
			jobDetails, err := c.JobDetailsClient.Get(ctx, vaultName, resourceGroupName, jobName, nil)
			if err != nil {
				return nil, false, err
			}
			storageJob, ok := jobDetails.Properties.(*armrecoveryservicesbackup.AzureStorageJob)
			if !ok {
				return nil, false, errors.New("failed to cast job details to AzureStorageJob")
			}
			if storageJob.Status != nil && *statusStr == string(armrecoveryservicesbackup.JobStatusInProgress) &&
				(storageJob.ExtendedInfo == nil || storageJob.ExtendedInfo.PropertyBag == nil || storageJob.ExtendedInfo.PropertyBag["RestoreDestination"] == nil) {
				retry = true
			}
			if storageJob.Status != nil && storageJob.ExtendedInfo != nil &&
				storageJob.ExtendedInfo.PropertyBag != nil && storageJob.ExtendedInfo.PropertyBag["RestoreDestination"] != nil &&
				strings.Contains(*storageJob.ExtendedInfo.PropertyBag["RestoreDestination"], restoreFolderPath) {
				return &jobName, false, nil
			}
		}
	}
	return nil, retry, nil
}

func (c jobsClient) GetStorageJob(ctx context.Context, vaultName string,
	resourceGroupName string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error) {
	job, err := c.JobDetailsClient.Get(ctx, vaultName, resourceGroupName, jobId, nil)
	if err != nil {
		return nil, err
	}
	jobDetails, ok := job.Properties.(*armrecoveryservicesbackup.AzureStorageJob)
	if !ok {
		return nil, errors.New("failed to cast job details to AzureStorageJob")
	}
	return jobDetails, nil
}
