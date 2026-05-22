package client

import (
	"context"
	"fmt"
	"sync"

	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"k8s.io/utils/ptr"
)

func Mock() awsclient.SkrClientProvider[NukeClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeClient, error) {
		backupClient, err := awsnfsvolumebackupclient.NewMockClient()(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return &mockClient{
			backupClient:  backupClient,
			webAcls:       make(map[string]*mockWebACL),
			certificates:  make(map[string]*mockCertificate),
			webAclsByName: make(map[string]string), // name -> ARN mapping
		}, nil
	}
}

type mockWebACL struct {
	Summary   wafv2types.WebACLSummary
	Tags      []wafv2types.Tag
	WebACL    *wafv2types.WebACL
	LockToken string
}

type mockCertificate struct {
	Summary acmtypes.CertificateSummary
	Tags    []acmtypes.Tag
}

type mockClient struct {
	backupClient awsnfsvolumebackupclient.Client

	mu            sync.Mutex
	webAcls       map[string]*mockWebACL      // ARN -> WebACL
	certificates  map[string]*mockCertificate // ARN -> Certificate
	webAclsByName map[string]string           // name -> ARN for lookup
}

// AddWebACL adds a WebACL to the mock for testing
func (m *mockClient) AddWebACL(name, id, arn string, tags []wafv2types.Tag) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.webAcls[arn] = &mockWebACL{
		Summary: wafv2types.WebACLSummary{
			ARN:  ptr.To(arn),
			Name: ptr.To(name),
			Id:   ptr.To(id),
		},
		Tags: tags,
		WebACL: &wafv2types.WebACL{
			ARN:  ptr.To(arn),
			Name: ptr.To(name),
			Id:   ptr.To(id),
		},
		LockToken: "mock-lock-token-" + id,
	}
	m.webAclsByName[name] = arn
}

// AddCertificate adds a Certificate to the mock for testing
func (m *mockClient) AddCertificate(arn string, tags []acmtypes.Tag) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.certificates[arn] = &mockCertificate{
		Summary: acmtypes.CertificateSummary{
			CertificateArn: ptr.To(arn),
		},
		Tags: tags,
	}
}

// IsWebACLDeleted checks if a WebACL was deleted (not found in map)
func (m *mockClient) IsWebACLDeleted(arn string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.webAcls[arn]
	return !exists
}

// IsCertificateDeleted checks if a Certificate was deleted (not found in map)
func (m *mockClient) IsCertificateDeleted(arn string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.certificates[arn]
	return !exists
}

// Implement NukeClient interface by embedding backup client methods
func (m *mockClient) IsNotFound(err error) bool {
	return m.backupClient.IsNotFound(err)
}

func (m *mockClient) IsAlreadyExists(err error) bool {
	return m.backupClient.IsAlreadyExists(err)
}

func (m *mockClient) ListTags(ctx context.Context, resourceArn string) (map[string]string, error) {
	return m.backupClient.ListTags(ctx, resourceArn)
}

func (m *mockClient) ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error) {
	return m.backupClient.ListBackupVaults(ctx)
}

func (m *mockClient) DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error) {
	return m.backupClient.DescribeBackupVault(ctx, backupVaultName)
}

func (m *mockClient) CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error) {
	return m.backupClient.CreateBackupVault(ctx, name, tags)
}

func (m *mockClient) DeleteBackupVault(ctx context.Context, name string) error {
	return m.backupClient.DeleteBackupVault(ctx, name)
}

func (m *mockClient) StartBackupJob(ctx context.Context, params *awsnfsvolumebackupclient.StartBackupJobInput) (*backup.StartBackupJobOutput, error) {
	return m.backupClient.StartBackupJob(ctx, params)
}

func (m *mockClient) DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error) {
	return m.backupClient.DescribeBackupJob(ctx, backupJobId)
}

func (m *mockClient) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]backuptypes.RecoveryPointByBackupVault, error) {
	return m.backupClient.ListRecoveryPointsForVault(ctx, accountId, backupVaultName)
}

func (m *mockClient) DescribeRecoveryPoint(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.DescribeRecoveryPointOutput, error) {
	return m.backupClient.DescribeRecoveryPoint(ctx, accountId, backupVaultName, recoveryPointArn)
}

func (m *mockClient) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	return m.backupClient.DeleteRecoveryPoint(ctx, backupVaultName, recoveryPointArn)
}

func (m *mockClient) StartCopyJob(ctx context.Context, params *awsnfsvolumebackupclient.StartCopyJobInput) (*backup.StartCopyJobOutput, error) {
	return m.backupClient.StartCopyJob(ctx, params)
}

func (m *mockClient) DescribeCopyJob(ctx context.Context, copyJobId string) (*backup.DescribeCopyJobOutput, error) {
	return m.backupClient.DescribeCopyJob(ctx, copyJobId)
}

// Mock WebACL methods
func (m *mockClient) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var summaries []wafv2types.WebACLSummary
	for _, wacl := range m.webAcls {
		summaries = append(summaries, wacl.Summary)
	}
	return summaries, nil
}

func (m *mockClient) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find by name
	if arn, ok := m.webAclsByName[name]; ok {
		if wacl, ok := m.webAcls[arn]; ok {
			return wacl.WebACL, wacl.LockToken, nil
		}
	}

	return nil, "", fmt.Errorf("WAFv2 WebACL not found")
}

func (m *mockClient) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find by name and delete from map
	if arn, ok := m.webAclsByName[name]; ok {
		if _, ok := m.webAcls[arn]; ok {
			delete(m.webAcls, arn)
			delete(m.webAclsByName, name)
			return nil
		}
	}

	return fmt.Errorf("WAFv2 WebACL not found")
}

func (m *mockClient) ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if wacl, ok := m.webAcls[resourceArn]; ok {
		return wacl.Tags, nil
	}
	return nil, nil
}

// Mock Certificate methods
func (m *mockClient) ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var summaries []acmtypes.CertificateSummary
	for _, cert := range m.certificates {
		summaries = append(summaries, cert.Summary)
	}
	return summaries, nil
}

func (m *mockClient) ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cert, ok := m.certificates[arn]; ok {
		return cert.Tags, nil
	}
	return nil, nil
}

func (m *mockClient) DeleteCertificate(ctx context.Context, arn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete from map
	if _, ok := m.certificates[arn]; ok {
		delete(m.certificates, arn)
		return nil
	}

	return fmt.Errorf("Certificate not found")
}
