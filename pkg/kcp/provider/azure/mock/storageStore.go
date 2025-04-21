package mock

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"k8s.io/utils/ptr"
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

	jobs                   map[string]*armrecoveryservicesbackup.JobDetailsClientGetResponse
	vaults                 []*armrecoveryservices.Vault
	protectedItems         map[string][]*armrecoveryservicesbackup.ProtectedItemResource
	backupProtectableItems []*armrecoveryservicesbackup.WorkloadProtectableItemResource
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
	s.m.Lock()
	defer s.m.Unlock()

	// TODO: create unhappy path?
	return s.backupProtectableItems, nil

}

func (s *storageStore) CreateBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {
	s.m.Lock()
	defer s.m.Unlock()

	// TODO: create unhappy path?
	return nil
}

func (s *storageStore) DeleteBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageStore) CreateOrUpdateProtectedItem(ctx context.Context, subscriptionId, location, vaultName, resourceGroupName, containerName, protectedItemName, backupPolicyName, storageAccountName string) error {
	s.m.Lock()
	defer s.m.Unlock()
	logger := composed.LoggerFromCtx(ctx)

	vaultId := client.GetVaultPath(s.subscription, resourceGroupName, vaultName)
	id := client.GetFileSharePath(s.subscription, resourceGroupName, vaultName, containerName, protectedItemName)
	protected := armrecoveryservicesbackup.ProtectedItemResource{
		Location: to.Ptr(location),
		ID:       to.Ptr(id),
		Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{
			ProtectionState: to.Ptr(armrecoveryservicesbackup.ProtectionStateProtected),
			FriendlyName:    to.Ptr(protectedItemName),
		},
	}

	//Add to the map.
	temp := s.protectedItems[vaultId]
	temp = append(temp, &protected)
	s.protectedItems[vaultId] = temp

	logger.Info("mock: Create/Update Protected Item", "protected-id", id)

	return nil
}

func (s *storageStore) RemoveProtection(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName string) error {
	s.m.Lock()
	defer s.m.Unlock()
	logger := composed.LoggerFromCtx(ctx)

	fileShareName := strings.TrimPrefix(protectedItemName, "AzureFileShare;")
	vaultId := client.GetVaultPath(s.subscription, resourceGroupName, vaultName)
	id := client.GetFileSharePath(s.subscription, resourceGroupName, vaultName, containerName, fileShareName)

	protectedItems, okay := s.protectedItems[vaultId]
	if !okay {
		return nil
	}

	temp := protectedItems[:0]
	for _, protected := range protectedItems {
		if ptr.Deref(protected.ID, "") != id {
			temp = append(temp, protected)
		}
	}
	s.protectedItems[vaultId] = temp

	logger.Info("mock: RemoveProtection", "protected-id", id)
	return nil
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
	s.m.Lock()
	defer s.m.Unlock()
	logger := composed.LoggerFromCtx(ctx)

	id := client.GetVaultPath(s.subscription, resourceGroupName, vaultName)
	vault := armrecoveryservices.Vault{
		Location: to.Ptr(location),
		ID:       to.Ptr(id),
		Name:     to.Ptr(vaultName),
		Tags: map[string]*string{
			"cloud-manager": to.Ptr("rwxVolumeBackup"),
		},
		Properties: &armrecoveryservices.VaultProperties{
			SecuritySettings: &armrecoveryservices.SecuritySettings{
				SoftDeleteSettings: &armrecoveryservices.SoftDeleteSettings{
					SoftDeleteState:                 ptr.To(armrecoveryservices.SoftDeleteStateEnabled),
					SoftDeleteRetentionPeriodInDays: ptr.To(int32(14)),
				},
			},
		},
	}
	s.vaults = append(s.vaults, &vault)

	logger.Info("mock: Create Vault", "vault-id", id)
	return vault.ID, nil
}

func (s *storageStore) DeleteVault(ctx context.Context, resourceGroupName string, vaultName string) error {
	s.m.Lock()
	defer s.m.Unlock()
	logger := composed.LoggerFromCtx(ctx)

	temp := s.vaults[:0]
	id := to.Ptr(client.GetVaultPath(s.subscription, resourceGroupName, vaultName))
	for _, vault := range s.vaults {
		if ptr.Deref(vault.ID, "") != ptr.Deref(id, "") {
			temp = append(temp, vault)
		}
	}
	s.vaults = temp

	logger.Info("mock: Delete Vault", "vault-id", id)
	return nil
}

func (s *storageStore) ListVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error) {
	s.m.Lock()
	defer s.m.Unlock()
	logger := composed.LoggerFromCtx(ctx)
	logger.Info("mock: List Vaults", "size", len(s.vaults))
	return s.vaults, nil
}

func (s *storageStore) TriggerBackup(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName, location string) error {
	s.m.Lock()
	defer s.m.Unlock()

	// TODO: create unhappy path?
	return nil

}

func (s *storageStore) ListProtectedItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {
	s.m.Lock()
	defer s.m.Unlock()

	vaultId := client.GetVaultPath(s.subscription, resourceGroupName, vaultName)
	items := s.protectedItems[vaultId]

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("mock: ListProtectedItems", "size", len(items))
	return items, nil
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

func (s *storageStore) GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error) {
	return &armrecoveryservicesbackup.BackupResourceVaultConfigResource{
		Properties: &armrecoveryservicesbackup.BackupResourceVaultConfig{
			EnhancedSecurityState:  to.Ptr(armrecoveryservicesbackup.EnhancedSecurityStateEnabled),
			SoftDeleteFeatureState: to.Ptr(armrecoveryservicesbackup.SoftDeleteFeatureStateEnabled),
		},
	}, nil
}
func (s *storageStore) PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error {
	return nil
}
func (s *storageStore) GetStorageContainers(ctx context.Context, resourceGroupName, vaultName string) ([]*armrecoveryservicesbackup.ProtectionContainerResource, error) {
	return nil, nil
}
func (s *storageStore) UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error {
	return nil
}

func newStorageStore(subscription string) *storageStore {
	return &storageStore{
		subscription:   subscription,
		jobs:           make(map[string]*armrecoveryservicesbackup.JobDetailsClientGetResponse),
		protectedItems: make(map[string][]*armrecoveryservicesbackup.ProtectedItemResource),
	}
}
