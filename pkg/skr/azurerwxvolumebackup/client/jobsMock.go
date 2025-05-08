package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"net/http"
	"time"
)

var (
	jobsMock *jobsMockClient
)

type restoreJob struct {
	jobId             string
	restoreFolderPath string
	status            string
	finalStatus       string
}

type backupJob struct {
	jobId       string
	status      string
	finalStatus string
	startTime   string
}

func newJobsMockClient() *jobsClient {
	return &jobsClient{}
}

type jobsMockClient struct {
	jobsClient
	restoreJobs []*restoreJob
	backupJobs  []*backupJob
}

func (m *jobsMockClient) FindRestoreJobId(_ context.Context, _ string, _ string, _ string, _ string, restoreFolderPath string) (*string, bool, error) {
	retry := false
	if restoreFolderPath == "" {
		return nil, false, errors.New("restoreFolderPath is empty")
	}
	for _, job := range m.restoreJobs {
		if job.restoreFolderPath == restoreFolderPath {
			if job.status != string(armrecoveryservicesbackup.JobStatusInProgress) {
				return &job.jobId, false, nil
			}
			if job.status == string(armrecoveryservicesbackup.JobStatusInProgress) {
				retry = true
				job.status = job.finalStatus
			}
		}
	}
	if retry {
		return nil, retry, nil
	}
	return nil, false, nil
}

func (m *jobsMockClient) FindNextBackupJobId(_ context.Context, _ string, _ string, _ string, startTime time.Time) (*string, error) {
	var results []string
	for _, job := range m.backupJobs {
		jobStartTime, err := time.Parse(time.RFC3339, job.startTime)
		if err != nil {
			return nil, err
		}
		if jobStartTime.After(startTime) {
			results = append(results, job.jobId)
		}
	}
	if len(results) > 1 {
		return nil, errors.New("multiple backup jobs found")
	}
	if len(results) == 1 {
		return &results[0], nil
	}
	return nil, nil
}

func (m *jobsMockClient) GetLastBackupJobStartTime(_ context.Context, _ string, _ string, _ string, _ time.Time) (*time.Time, error) {
	var lastStartTime *time.Time
	for _, job := range m.backupJobs {
		startTime, err := time.Parse(time.RFC3339, job.startTime)
		if err != nil {
			return nil, err
		}
		if lastStartTime == nil || startTime.After(*lastStartTime) {
			lastStartTime = &startTime
		}
	}
	return lastStartTime, nil
}

func (m *jobsMockClient) GetStorageJob(_ context.Context, _ string,
	_ string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error) {
	for _, job := range jobsMock.restoreJobs {
		if job.jobId == jobId {
			status := job.status
			if job.status == string(armrecoveryservicesbackup.JobStatusInProgress) {
				job.status = job.finalStatus
			}
			return to.Ptr(armrecoveryservicesbackup.AzureStorageJob{
				Status: to.Ptr(status),
			}), nil
		}
	}
	for _, job := range jobsMock.backupJobs {
		if job.jobId == jobId {
			if job.status == string(armrecoveryservicesbackup.JobStatusInProgress) {
				job.status = job.finalStatus
			}
			return to.Ptr(armrecoveryservicesbackup.AzureStorageJob{
				Status: to.Ptr(job.status),
			}), nil
		}
	}
	err := runtime.NewResponseError(&http.Response{
		StatusCode: http.StatusNotFound,
	})
	return nil, err
}

func (m *jobsMockClient) AddRestoreJob(jobId, restoreFolderPath string, status, finalStatus armrecoveryservicesbackup.JobStatus) {
	m.restoreJobs = append(m.restoreJobs, &restoreJob{
		jobId:             jobId,
		restoreFolderPath: restoreFolderPath,
		status:            string(status),
		finalStatus:       string(finalStatus),
	})
}

func (m *jobsMockClient) AddBackupJob(jobId string, status, finalStatus armrecoveryservicesbackup.JobStatus, startTime string) {
	m.backupJobs = append(m.backupJobs, &backupJob{
		jobId:       jobId,
		status:      string(status),
		finalStatus: string(finalStatus),
		startTime:   startTime,
	})
}
