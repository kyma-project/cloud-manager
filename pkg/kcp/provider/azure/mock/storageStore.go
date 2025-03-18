package mock

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"strings"
	"sync"
	"time"
)

var _ client.RestoreClient = &storageStore{}
var _ client.BackupClient = &storageStore{}
var _ client.VaultClient = &storageStore{}
var _ client.JobsClient = &storageStore{}
var _ client.RecoveryPointClient = &storageStore{}
var _ client.ProtectedItemsClient = &storageStore{}
var _ client.ProtectionPoliciesClient = &storageStore{}
var _ client.BackupProtectableItemsClient = &storageStore{}

type storageStore struct {
	m            sync.Mutex
	subscription string

	jobs map[string]*armrecoveryservicesbackup.JobDetailsClientGetResponse
}

func (s *storageStore) FindRestoreJobId(ctx context.Context, vaultName string, resourceGroupName string, fileShareName string, startFilter string, restoreFolderPath string) (*string, bool, error) {
	if isContextCanceled(ctx) {
		return nil, false, errors.New("context canceled")
	}
	s.m.Lock()
	defer s.m.Unlock()
	idPathPrefix := client.GetVaultPath(s.subscription, resourceGroupName, vaultName) + "/backupJobs/"
	for _, job := range s.jobs {
		if *job.Properties.GetJob().EntityFriendlyName == fileShareName && strings.Contains(*job.ID, idPathPrefix) {
			expectedStartTimeFilter, err := time.Parse(time.RFC3339, startFilter)
			if err != nil {
				return nil, false, err
			}
			if !job.Properties.GetJob().StartTime.After(expectedStartTimeFilter) {
				continue
			}
			storageJob, _ := job.Properties.(*armrecoveryservicesbackup.AzureStorageJob)
			if storageJob.EntityFriendlyName == nil || *storageJob.EntityFriendlyName != fileShareName {
				continue
			}
			if storageJob.ExtendedInfo.PropertyBag == nil || storageJob.ExtendedInfo.PropertyBag["RestoreDestination"] == nil {
				return nil, true, nil
			}
			if strings.Contains(*storageJob.ExtendedInfo.PropertyBag["RestoreDestination"], restoreFolderPath) {
				return job.Name, false, nil
			}
		}
	}
	return nil, false, nil
}

func (s *storageStore) changeStatusToCompletedDelayed(ctx context.Context, storageJob *armrecoveryservicesbackup.AzureStorageJob, duration time.Duration) {
	time.Sleep(duration)
	s.m.Lock()
	defer s.m.Unlock()
	storageJob.Status = to.Ptr(string(armrecoveryservicesbackup.JobStatusCompleted))
	storageJob.EndTime = to.Ptr(time.Now())
}
func (s *storageStore) GetStorageJob(ctx context.Context, _ string, _ string, jobId string) (*armrecoveryservicesbackup.AzureStorageJob, error) {
	if isContextCanceled(ctx) {
		return nil, errors.New("context canceled")
	}
	s.m.Lock()
	defer s.m.Unlock()
	job, ok := s.jobs[jobId]
	if !ok {
		return nil, errors.New("job not found")
	}
	storageJob, ok := job.Properties.(*armrecoveryservicesbackup.AzureStorageJob)
	if !ok {
		return nil, errors.New("failed to cast job details to AzureStorageJob")
	}
	go s.changeStatusToCompletedDelayed(ctx, storageJob, 3*time.Second)
	return storageJob, nil
}

func (s *storageStore) ListBackupProtectableItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.WorkloadProtectableItemResource, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) CreateBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) DeleteBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) CreateOrUpdateProtectedItem(ctx context.Context, subscriptionId, location, vaultName, resourceGroupName, containerName, protectedItemName, backupPolicyName, storageAccountName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) GetRecoveryPoint(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string, recoveryPointId string) (armrecoveryservicesbackup.RecoveryPointResource, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) ListRecoveryPoints(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string) ([]*armrecoveryservicesbackup.RecoveryPointResource, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) CreateVault(ctx context.Context, resourceGroupName string, vaultName string, location string) (*string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) DeleteVault(ctx context.Context, resourceGroupName string, vaultName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) ListVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) TriggerBackup(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName, location string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) ListProtectedItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) TriggerRestore(ctx context.Context, request client.RestoreRequest) (*string, error) {
	if isContextCanceled(ctx) {
		return nil, errors.New("context canceled")
	}
	s.m.Lock()
	defer s.m.Unlock()
	jobId := uuid.NewString()
	idPath := client.GetVaultPath(s.subscription, request.ResourceGroupName, request.VaultName) + "/backupJobs/" + jobId
	inProgress := string(armrecoveryservicesbackup.JobStatusInProgress)
	JobDetailsClientGetResponse := armrecoveryservicesbackup.JobDetailsClientGetResponse{
		JobResource: armrecoveryservicesbackup.JobResource{
			ID:   &idPath,
			Name: &jobId,
			Properties: to.Ptr(armrecoveryservicesbackup.AzureStorageJob{
				Status:             &inProgress,
				EntityFriendlyName: to.Ptr(request.TargetFileShareName),
				ExtendedInfo: to.Ptr(armrecoveryservicesbackup.AzureStorageJobExtendedInfo{
					PropertyBag: map[string]*string{
						"RestoreDestination": to.Ptr(request.TargetFolderName),
					},
				}),
				StartTime: to.Ptr(time.Now()),
			}),
		},
	}
	s.jobs[jobId] = &JobDetailsClientGetResponse
	return &jobId, nil
}

func newStorageStore(subscription string) *storageStore {
	return &storageStore{
		subscription: subscription,
		jobs:         make(map[string]*armrecoveryservicesbackup.JobDetailsClientGetResponse),
	}
}
