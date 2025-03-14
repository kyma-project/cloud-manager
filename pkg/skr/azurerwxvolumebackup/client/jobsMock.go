package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

var (
	jobsMock *jobsMockClient
)

type restoreJob struct {
	jobId             string
	restoreFolderPath string
	status            string
}

func newJobsMockClient() *jobsClient {
	return &jobsClient{}
}

type jobsMockClient struct {
	jobsClient
	restoreJobs []*restoreJob
}

func (m *jobsMockClient) FindRestoreJobId(_ context.Context, _ string, _ string, _ string, _ string, restoreFolderPath string) (*string, bool, error) {
	retry := false
	if restoreFolderPath == "" {
		return nil, false, errors.New("restoreFolderPath is empty")
	}
	for _, job := range m.restoreJobs {
		if job.restoreFolderPath == restoreFolderPath {
			if job.status == string(armrecoveryservicesbackup.JobStatusCompleted) {
				return &job.jobId, false, nil
			}
			if job.status == string(armrecoveryservicesbackup.JobStatusInProgress) {
				retry = true
				job.status = string(armrecoveryservicesbackup.JobStatusCompleted)
			}
		}
	}
	if retry {
		return nil, retry, nil
	}
	return nil, false, nil
}

func (m *jobsMockClient) GetStorageJob(_ context.Context, _ string,
	_ string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error) {
	for _, job := range jobsMock.restoreJobs {
		if job.jobId == jobId {
			return to.Ptr(armrecoveryservicesbackup.AzureStorageJob{
				Status: to.Ptr(job.status),
			}), nil
		}
	}
	return nil, errors.New("job not found")
}

func (m *jobsMockClient) AddStorageJob(jobId, restoreFolderPath string, status armrecoveryservicesbackup.JobStatus) {
	m.restoreJobs = append(m.restoreJobs, &restoreJob{
		jobId:             jobId,
		restoreFolderPath: restoreFolderPath,
		status:            string(status),
	})
}
