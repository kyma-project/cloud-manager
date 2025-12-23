package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"strings"
	"time"
)

type JobsClient interface {
	FindRestoreJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startFilter string, restoreFolderPath string) (*string, bool, error)
	GetStorageJob(ctx context.Context, vaultName string, resourceGroupName string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error)
	GetLastBackupJobStartTime(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startTime time.Time) (*time.Time, error)
	FindNextBackupJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startTime time.Time) (*string, error)
}
type jobsClient struct {
	*armrecoveryservicesbackup.BackupJobsClient
	*armrecoveryservicesbackup.JobDetailsClient
}

func NewJobsClient(bjc *armrecoveryservicesbackup.BackupJobsClient, jdc *armrecoveryservicesbackup.JobDetailsClient) JobsClient {
	return jobsClient{bjc, jdc}
}

// FindRestoreJobId finds the restore job for the given vault, resource group, start filter and restore folder path
// If a job is in progress and the destination folder is not populated yet, it returns true as second return value which indicates that this should be retried at a later time
func (c jobsClient) FindRestoreJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startFilter string, restoreFolderPath string) (*string, bool, error) {
	backupJobsClientOptions := armrecoveryservicesbackup.BackupJobsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("backupManagementType eq 'AzureStorage' and operation eq 'Restore' and startTime ge %v", startFilter)),
	}
	pager := c.NewListPager(vaultName, resourceGroupName, &backupJobsClientOptions)
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
			jobDetails, err := c.Get(ctx, vaultName, resourceGroupName, jobName, nil)
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
	job, err := c.Get(ctx, vaultName, resourceGroupName, jobId, nil)
	if err != nil {
		return nil, err
	}
	jobDetails, ok := job.Properties.(*armrecoveryservicesbackup.AzureStorageJob)
	if !ok {
		return nil, errors.New("failed to cast job details to AzureStorageJob")
	}
	return jobDetails, nil
}

// GetLastBackupJobStartTime returns the start time of the last backup job for the given vault, resource group and file share name
func (c jobsClient) GetLastBackupJobStartTime(ctx context.Context, vaultName string,
	resourceGroupName string, fileShareName string, startTime time.Time) (*time.Time, error) {
	var lastStartTime *time.Time
	backupJobsClientOptions := armrecoveryservicesbackup.BackupJobsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("backupManagementType eq 'AzureStorage' and operation eq 'Backup' and startTime eq '%v'", ToStorageJobTimeFilter(startTime))),
	}
	pager := c.NewListPager(vaultName, resourceGroupName, &backupJobsClientOptions)
	if pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting next page: %w", err)
		}
		for _, job := range page.Value {
			if job.Properties.GetJob().EntityFriendlyName == nil || *job.Properties.GetJob().EntityFriendlyName != fileShareName {
				continue
			}
			startTime := job.Properties.GetJob().StartTime
			if lastStartTime == nil || (startTime != nil && lastStartTime.Before(*startTime)) {
				lastStartTime = startTime
			}
		}
	}
	if lastStartTime == nil {
		return nil, nil
	}
	return lastStartTime, nil
}

// FindNextBackupJobId finds the backup job for the given vault, resource group, file share name that started after given startTime
// It will error out, if it finds multiple jobs
func (c jobsClient) FindNextBackupJobId(ctx context.Context, vaultName string,
	resourceGroupName string, fileShareName string, startTime time.Time) (*string, error) {
	backupJobsClientOptions := armrecoveryservicesbackup.BackupJobsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("backupManagementType eq 'AzureStorage' "+
			"and operation eq 'Backup' and startTime eq '%v'", ToStorageJobTimeFilter(startTime))),
	}
	var jobIds []string
	pager := c.NewListPager(vaultName, resourceGroupName, &backupJobsClientOptions)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting next page: %w", err)
		}
		for _, job := range page.Value {
			if job.Properties.GetJob().EntityFriendlyName == nil || *job.Properties.GetJob().EntityFriendlyName != fileShareName {
				continue
			}
			if job.Properties.GetJob().StartTime.Equal(startTime) {
				continue
			}
			jobIds = append(jobIds, *job.Name)
		}
	}
	if len(jobIds) == 1 {
		return &jobIds[0], nil
	}
	if len(jobIds) > 1 {
		return nil, fmt.Errorf("multiple backup jobs found for file share %s after specified start time of %s", fileShareName, startTime)
	}
	return nil, nil
}
