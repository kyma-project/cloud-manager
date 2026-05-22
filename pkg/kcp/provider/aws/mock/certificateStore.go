package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

type CertificateConfig interface {
	SetCertificateError(arn string, err error)
	InitiateCertificate(arn string, cert []byte, key []byte)
	GetCertificateByArn(arn string) *acmtypes.CertificateDetail
	GetCertificateTags(arn string) []acmtypes.Tag
	SetCertificateInUse(arn string, inUse bool)
}

type certificateEntry struct {
	detail           *acmtypes.CertificateDetail
	inUse            bool
	certificate      []byte         // The actual certificate PEM
	certificateChain []byte         // The certificate chain PEM
	tags             []acmtypes.Tag // Tags associated with the certificate
}

type certificateStore struct {
	m        sync.Mutex
	items    map[string]*certificateEntry // ARN -> entry
	errorMap map[string]error
	account  string
	region   string
}

func newCertificateStore(account, region string) *certificateStore {
	return &certificateStore{
		items:    make(map[string]*certificateEntry),
		errorMap: make(map[string]error),
		account:  account,
		region:   region,
	}
}

func (s *certificateStore) ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var arn string
	var entry *certificateEntry

	// If ARN is provided, update existing certificate
	if input.CertificateArn != nil {
		arn = ptr.Deref(input.CertificateArn, "")
		if err, ok := s.errorMap[arn]; ok && err != nil {
			return "", err
		}

		var ok bool
		entry, ok = s.items[arn]
		if !ok {
			return "", &smithy.GenericAPIError{
				Code:    "ResourceNotFoundException",
				Message: fmt.Sprintf("Certificate with ARN %s does not exist", arn),
			}
		}
	} else {
		// Create new certificate
		id := uuid.NewString()
		arn = awsutil.CertificateArn(s.region, s.account, id)
		entry = &certificateEntry{
			detail: &acmtypes.CertificateDetail{
				CertificateArn: ptr.To(arn),
				DomainName:     ptr.To("example.com"), // Parse from certificate in real impl
				Status:         acmtypes.CertificateStatusIssued,
				Type:           acmtypes.CertificateTypeImported,
				CreatedAt:      ptr.To(time.Now()),
			},
		}
		s.items[arn] = entry
	}

	// Update certificate data
	entry.detail.ImportedAt = ptr.To(time.Now())
	entry.detail.NotBefore = ptr.To(time.Now())
	entry.detail.NotAfter = ptr.To(time.Now().Add(365 * 24 * time.Hour)) // 1 year expiration

	// Store the certificate and chain data
	entry.certificate = input.Certificate
	if input.CertificateChain != nil {
		entry.certificateChain = input.CertificateChain
	}

	// Store tags if provided
	if input.Tags != nil {
		entry.tags = input.Tags
	}

	// Deep copy to avoid shared references
	detailCopy, err := util.JsonClone(entry.detail)
	if err != nil {
		return "", err
	}
	entry.detail = detailCopy

	return arn, nil
}

func (s *certificateStore) DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[arn]; ok && err != nil {
		return nil, err
	}

	entry, ok := s.items[arn]
	if !ok {
		return nil, &smithy.GenericAPIError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Certificate with ARN %s does not exist", arn),
		}
	}

	detailCopy, err := util.JsonClone(entry.detail)
	if err != nil {
		return nil, err
	}

	return detailCopy, nil
}

func (s *certificateStore) DeleteCertificate(ctx context.Context, arn string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[arn]; ok && err != nil {
		return err
	}

	entry, ok := s.items[arn]
	if !ok {
		return &smithy.GenericAPIError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Certificate with ARN %s does not exist", arn),
		}
	}

	// Check if certificate is in use
	if entry.inUse {
		return &smithy.GenericAPIError{
			Code:    "ResourceInUseException",
			Message: "Certificate is currently in use by other AWS resources",
		}
	}

	delete(s.items, arn)
	return nil
}

func (s *certificateStore) GetCertificate(ctx context.Context, arn string) (string, string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[arn]; ok && err != nil {
		return "", "", err
	}

	entry, ok := s.items[arn]
	if !ok {
		return "", "", &smithy.GenericAPIError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("Certificate with ARN %s does not exist", arn),
		}
	}

	return string(entry.certificate), string(entry.certificateChain), nil
}

func (s *certificateStore) SetCertificateError(arn string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.errorMap[arn] = err
}

func (s *certificateStore) InitiateCertificate(arn string, cert []byte, key []byte) {
	s.m.Lock()
	defer s.m.Unlock()

	entry := &certificateEntry{
		detail: &acmtypes.CertificateDetail{
			CertificateArn: ptr.To(arn),
			DomainName:     ptr.To("example.com"),
			Status:         acmtypes.CertificateStatusIssued,
			Type:           acmtypes.CertificateTypeImported,
			CreatedAt:      ptr.To(time.Now()),
			ImportedAt:     ptr.To(time.Now()),
			NotBefore:      ptr.To(time.Now()),
			NotAfter:       ptr.To(time.Now().Add(365 * 24 * time.Hour)),
		},
		certificate: cert,
	}

	s.items[arn] = entry
}

func (s *certificateStore) GetCertificateByArn(arn string) *acmtypes.CertificateDetail {
	s.m.Lock()
	defer s.m.Unlock()

	entry, ok := s.items[arn]
	if !ok {
		return nil
	}

	detailCopy, _ := util.JsonClone(entry.detail)
	return detailCopy
}

func (s *certificateStore) SetCertificateInUse(arn string, inUse bool) {
	s.m.Lock()
	defer s.m.Unlock()

	if entry, ok := s.items[arn]; ok {
		entry.inUse = inUse
	}
}

func (s *certificateStore) GetCertificateTags(arn string) []acmtypes.Tag {
	s.m.Lock()
	defer s.m.Unlock()

	entry, ok := s.items[arn]
	if !ok {
		return nil
	}

	// Return a copy of tags to avoid external modifications
	tagsCopy := make([]acmtypes.Tag, len(entry.tags))
	copy(tagsCopy, entry.tags)
	return tagsCopy
}

func (s *certificateStore) AcmClient() awsclient.AcmClient {
	return s
}
