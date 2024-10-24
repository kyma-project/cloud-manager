package client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
	"time"
)

var (
	MockAwsAccount = "some-aws-account"

	mock = &mockClient{
		localClient: *newLocalClient(),
	}
)

type mockClient struct {
	localClient
	restoreJobs []backup.DescribeRestoreJobOutput
}

func (m *mockClient) StartRestoreJob(ctx context.Context, params *StartRestoreJobInput) (*backup.StartRestoreJobOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	jobId := uuid.NewString()
	restoreJobOutput := backup.DescribeRestoreJobOutput{
		AccountId:        ptr.To(MockAwsAccount),
		RestoreJobId:     ptr.To(jobId),
		CreationDate:     ptr.To(time.Now()),
		IamRoleArn:       ptr.To(params.IamRoleArn),
		RecoveryPointArn: params.RecoveryPointArn,
		ResourceType:     nil,
		Status:           backuptypes.RestoreJobStatusCompleted,
	}
	m.restoreJobs = append(m.restoreJobs, restoreJobOutput)

	logger.WithName("StartRestoreJob - mock").Info(
		fmt.Sprintf("Restore Job ID :: %s, RecoveryPointArn :: %s", jobId, *params.RecoveryPointArn))
	return &backup.StartRestoreJobOutput{
		RestoreJobId: ptr.To(jobId),
	}, nil
}

func (m *mockClient) DescribeRestoreJob(_ context.Context, restoreJobId string) (*backup.DescribeRestoreJobOutput, error) {
	for _, job := range m.restoreJobs {
		if ptr.Deref(job.RestoreJobId, "") == restoreJobId {
			return &job, nil
		}
	}
	return nil, &backuptypes.ResourceNotFoundException{
		Message: ptr.To("BackupJob not found"),
	}
}

func (m *mockClient) GetRecoveryPointRestoreMetadata(_ context.Context, _, backupVaultName, recoveryPointArn string) (*backup.GetRecoveryPointRestoreMetadataOutput, error) {
	return &backup.GetRecoveryPointRestoreMetadataOutput{
		BackupVaultArn:   ptr.To(backupVaultName),
		RecoveryPointArn: ptr.To(recoveryPointArn),
		RestoreMetadata:  map[string]string{},
		ResourceType:     ptr.To("EFS"),
	}, nil
}

func NewMockClient() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		return mock, nil
	}
}
