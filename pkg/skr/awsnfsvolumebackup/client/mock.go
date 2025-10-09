package client

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/utils/ptr"
)

var (
	MockAwsRegion  = "eu-west-1"
	MockAwsAccount = "some-aws-account"

	clients = make(map[string]*mockClient)
)

type mockClient struct {
	localClient
	account        string
	location       string
	vaults         []backup.DescribeBackupVaultOutput
	backupJobs     []backup.DescribeBackupJobOutput
	recoveryPoints []backup.DescribeRecoveryPointOutput
	tags           map[string]map[string]string
	copyJobs       []backup.DescribeCopyJobOutput
}

func NewMockClient() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		return getMockClient(account, region)
	}
}

func getMockClient(account, region string) (*mockClient, error) {
	clientKey := fmt.Sprintf("%s::%s", account, region)
	client, exists := clients[clientKey]
	if !exists {
		client = &mockClient{
			localClient: *newLocalClient(),
			account:     account,
			location:    region,
			tags:        make(map[string]map[string]string),
		}
		clients[clientKey] = client
	}
	return client, nil
}

func (s *mockClient) ListTags(ctx context.Context, resourceArn string) (map[string]string, error) {
	return s.tags[resourceArn], nil
}

func (s *mockClient) ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error) {
	logger := composed.LoggerFromCtx(ctx)
	var vaultsList []backuptypes.BackupVaultListMember
	for _, vault := range s.vaults {
		vaultsList = append(vaultsList, backuptypes.BackupVaultListMember{
			BackupVaultArn:         vault.BackupVaultArn,
			BackupVaultName:        vault.BackupVaultName,
			CreationDate:           vault.CreationDate,
			CreatorRequestId:       vault.CreatorRequestId,
			EncryptionKeyArn:       vault.EncryptionKeyArn,
			LockDate:               vault.LockDate,
			Locked:                 vault.Locked,
			MaxRetentionDays:       vault.MaxRetentionDays,
			MinRetentionDays:       vault.MinRetentionDays,
			NumberOfRecoveryPoints: vault.NumberOfRecoveryPoints,
		})
	}
	logger.WithName("ListBackupVaults - mock").Info(fmt.Sprintf("Length :: %d", len(vaultsList)))

	return vaultsList, nil
}
func (s *mockClient) DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	for _, vault := range s.vaults {
		if ptr.Deref(vault.BackupVaultName, "") == backupVaultName {
			logger.Info(fmt.Sprintf("Returing vault: %s", ptr.Deref(vault.BackupVaultName, "")))
			return &vault, nil
		}
	}
	logger.Info(fmt.Sprintf("Vault not found [%s]", backupVaultName))
	return nil, &backuptypes.ResourceNotFoundException{
		Message: ptr.To("Vault not found"),
	}
}
func (s *mockClient) CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)
	arnTxt := awsutil.BackupVaultArn(s.location, s.account, name)
	vault := backup.DescribeBackupVaultOutput{
		BackupVaultArn:         ptr.To(arnTxt),
		BackupVaultName:        ptr.To(name),
		CreationDate:           ptr.To(time.Now()),
		NumberOfRecoveryPoints: 0,
	}
	s.vaults = append(s.vaults, vault)
	s.tags[arnTxt] = tags
	logger.WithName("CreateBackupVault - mock").Info(fmt.Sprintf("Length :: %d", len(s.vaults)))
	return ptr.Deref(vault.BackupVaultArn, ""), nil
}
func (s *mockClient) DeleteBackupVault(ctx context.Context, name string) error {
	logger := composed.LoggerFromCtx(ctx)
	for i, vault := range s.vaults {
		if ptr.Deref(vault.BackupVaultName, "") == name {
			s.vaults = append(s.vaults[:i], s.vaults[i+1:]...)
			logger.WithName("DeleteBackupVault - mock").Info(fmt.Sprintf("Length :: %d", len(s.vaults)))
		}
	}
	logger.WithName("DeleteBackupVault - mock").Info(fmt.Sprintf("Length :: %d", len(s.vaults)))
	return nil
}
func (s *mockClient) StartBackupJob(ctx context.Context, params *StartBackupJobInput) (*backup.StartBackupJobOutput, error) {

	logger := composed.LoggerFromCtx(ctx)
	vault, err := s.DescribeBackupVault(ctx, params.BackupVaultName)
	if err != nil {
		return nil, err
	}
	rPointId := uuid.NewString()
	jobId := uuid.NewString()

	arnTxt := awsutil.BackupRecoveryPointArn(s.location, s.account, rPointId)
	rPoint := backup.DescribeRecoveryPointOutput{
		BackupVaultArn:    vault.BackupVaultArn,
		BackupVaultName:   vault.BackupVaultName,
		CreationDate:      ptr.To(time.Now()),
		IamRoleArn:        ptr.To(params.IamRoleArn),
		IsEncrypted:       false,
		RecoveryPointArn:  ptr.To(arnTxt),
		ResourceArn:       ptr.To(params.ResourceArn),
		ResourceType:      nil,
		Status:            backuptypes.RecoveryPointStatusCompleted,
		BackupSizeInBytes: ptr.To(rand.Int64N(10240)),
	}

	job := backup.DescribeBackupJobOutput{
		AccountId:        ptr.To(s.account),
		BackupJobId:      ptr.To(jobId),
		BackupVaultArn:   rPoint.BackupVaultArn,
		BackupVaultName:  rPoint.BackupVaultName,
		CreationDate:     rPoint.CreationDate,
		IamRoleArn:       rPoint.IamRoleArn,
		RecoveryPointArn: rPoint.RecoveryPointArn,
		ResourceArn:      rPoint.ResourceArn,
		State:            backuptypes.BackupJobStateCompleted,
	}
	s.backupJobs = append(s.backupJobs, job)
	s.recoveryPoints = append(s.recoveryPoints, rPoint)
	logger.WithName("StartBackupJob - mock").Info(
		fmt.Sprintf("Backup ID :: %s, RecoveryPointArn :: %s", jobId, arnTxt))

	return &backup.StartBackupJobOutput{
		BackupJobId:      job.BackupJobId,
		CreationDate:     job.CreationDate,
		IsParent:         false,
		RecoveryPointArn: rPoint.RecoveryPointArn,
	}, nil
}
func (s *mockClient) DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error) {
	for _, job := range s.backupJobs {
		if ptr.Deref(job.BackupJobId, "") == backupJobId {
			return &job, nil
		}
	}
	return nil, &backuptypes.ResourceNotFoundException{
		Message: ptr.To("BackupJob not found"),
	}
}

func (s *mockClient) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]backuptypes.RecoveryPointByBackupVault, error) {
	var result []backuptypes.RecoveryPointByBackupVault
	for _, rp := range s.recoveryPoints {
		if ptr.Deref(rp.BackupVaultName, "") == backupVaultName {
			result = append(result, backuptypes.RecoveryPointByBackupVault{
				BackupSizeInBytes: rp.BackupSizeInBytes,
				BackupVaultArn:    rp.BackupVaultArn,
				BackupVaultName:   rp.BackupVaultName,
				IamRoleArn:        rp.IamRoleArn,
				RecoveryPointArn:  rp.RecoveryPointArn,
				ResourceArn:       rp.ResourceArn,
				ResourceName:      rp.ResourceName,
				ResourceType:      rp.ResourceType,
				Status:            rp.Status,
				StatusMessage:     rp.StatusMessage,
			})
		}
	}
	return result, nil
}

func (s *mockClient) DescribeRecoveryPoint(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.DescribeRecoveryPointOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("DescribeRecoveryPoint - mock").Info(
		fmt.Sprintf("RecoveryPointArn :: %s", recoveryPointArn))
	for _, rPoint := range s.recoveryPoints {
		if ptr.Deref(rPoint.RecoveryPointArn, "") == recoveryPointArn {
			return &rPoint, nil
		}
	}
	return nil, &backuptypes.ResourceNotFoundException{
		Message: ptr.To("RecoveryPoint not found"),
	}
}
func (s *mockClient) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	for i, rPoint := range s.recoveryPoints {
		if ptr.Deref(rPoint.RecoveryPointArn, "") == recoveryPointArn {
			s.recoveryPoints = append(s.recoveryPoints[:i], s.recoveryPoints[i+1:]...)
			logger.WithName("DeleteRecoveryPoint - mock").Info(fmt.Sprintf("Length :: %d", len(s.vaults)))
		}
	}
	return &backup.DeleteRecoveryPointOutput{}, nil
}

func (s *mockClient) StartCopyJob(ctx context.Context, params *StartCopyJobInput) (*backup.StartCopyJobOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	dstArnObj, err := arn.Parse(params.DestinationBackupVaultArn)
	if err != nil {
		return nil, fmt.Errorf("error parsing DestinationBackupVaultArn: %w", err)
	}
	vaultName := awsutil.StripArnResourceType(dstArnObj.Resource)

	srcVault, err := s.DescribeBackupVault(ctx, params.SourceBackupVaultName)
	if err != nil {
		return nil, err
	}

	//Source RecoveryPoint
	srcRecoveryPoint, err := s.DescribeRecoveryPoint(ctx, dstArnObj.AccountID, vaultName, params.RecoveryPointArn)
	if err != nil {
		return nil, err
	}

	//Destination client
	dstClient, err := getMockClient(dstArnObj.AccountID, dstArnObj.Region)
	if err != nil {
		return nil, err
	}

	//Destination Vault
	dstVault, err := dstClient.DescribeBackupVault(ctx, vaultName)
	if err != nil {
		return nil, err
	}
	rPointId := uuid.NewString()
	jobId := uuid.NewString()

	dstArn := awsutil.BackupRecoveryPointArn(dstArnObj.Region, dstArnObj.AccountID, rPointId)
	rPoint := backup.DescribeRecoveryPointOutput{
		BackupVaultArn:         dstVault.BackupVaultArn,
		BackupVaultName:        dstVault.BackupVaultName,
		CreationDate:           ptr.To(time.Now()),
		IamRoleArn:             ptr.To(params.IamRoleArn),
		IsEncrypted:            false,
		RecoveryPointArn:       ptr.To(dstArn),
		ResourceType:           nil,
		Status:                 backuptypes.RecoveryPointStatusCompleted,
		BackupSizeInBytes:      ptr.To(rand.Int64N(10240)),
		ParentRecoveryPointArn: srcRecoveryPoint.RecoveryPointArn,
		ResourceArn:            srcRecoveryPoint.ResourceArn,
		ResourceName:           srcRecoveryPoint.ResourceName,
		SourceBackupVaultArn:   srcVault.BackupVaultArn,
		VaultType:              dstVault.VaultType,
	}

	job := backup.DescribeCopyJobOutput{
		CopyJob: &backuptypes.CopyJob{
			AccountId:                   ptr.To(s.account),
			CreationDate:                rPoint.CreationDate,
			IamRoleArn:                  rPoint.IamRoleArn,
			ResourceArn:                 rPoint.ResourceArn,
			State:                       backuptypes.CopyJobStateCompleted,
			BackupSizeInBytes:           ptr.To(rand.Int64N(10240)),
			CompletionDate:              rPoint.CreationDate,
			CopyJobId:                   ptr.To(jobId),
			CreatedBy:                   &backuptypes.RecoveryPointCreator{},
			DestinationBackupVaultArn:   srcVault.BackupVaultArn,
			DestinationRecoveryPointArn: rPoint.RecoveryPointArn,
			IsParent:                    false,
		},
	}
	s.copyJobs = append(s.copyJobs, job)
	dstClient.recoveryPoints = append(dstClient.recoveryPoints, rPoint)
	logger.WithName("StartCopyJob - mock").Info(
		fmt.Sprintf("Backup ID :: %s, RecoveryPointArn :: %s", jobId, dstArn))

	return &backup.StartCopyJobOutput{
		CopyJobId:    job.CopyJob.CopyJobId,
		CreationDate: &time.Time{},
		IsParent:     false,
	}, nil
}

func (s *mockClient) DescribeCopyJob(ctx context.Context, copyJobId string) (*backup.DescribeCopyJobOutput, error) {
	for _, job := range s.copyJobs {
		if ptr.Deref(job.CopyJob.CopyJobId, "") == copyJobId {
			return &job, nil
		}
	}
	return nil, &backuptypes.ResourceNotFoundException{
		Message: ptr.To("CopyJob not found"),
	}
}
