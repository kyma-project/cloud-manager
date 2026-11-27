package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"k8s.io/utils/ptr"
)

type AcmClient interface {
	ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error)
	DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error)
	GetCertificate(ctx context.Context, arn string) (certificate string, certificateChain string, err error)
	DeleteCertificate(ctx context.Context, arn string) error
	ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error)
	ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error)
	ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error)
}

func NewAcmClient(svc *acm.Client) AcmClient {
	return &acmClient{
		svc: svc,
	}
}

var _ AcmClient = (*acmClient)(nil)

type acmClient struct {
	svc *acm.Client
}

func (c *acmClient) ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error) {
	out, err := c.svc.ImportCertificate(ctx, input)
	if err != nil {
		return "", err
	}
	return ptr.Deref(out.CertificateArn, ""), nil
}

func (c *acmClient) DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error) {
	in := &acm.DescribeCertificateInput{
		CertificateArn: new(arn),
	}

	out, err := c.svc.DescribeCertificate(ctx, in)
	if err != nil {
		return nil, err
	}

	return out.Certificate, nil
}

func (c *acmClient) DeleteCertificate(ctx context.Context, arn string) error {
	in := &acm.DeleteCertificateInput{
		CertificateArn: new(arn),
	}

	_, err := c.svc.DeleteCertificate(ctx, in)
	return err
}

func (c *acmClient) GetCertificate(ctx context.Context, arn string) (string, string, error) {
	in := &acm.GetCertificateInput{
		CertificateArn: new(arn),
	}

	out, err := c.svc.GetCertificate(ctx, in)
	if err != nil {
		return "", "", err
	}

	certificate := ptr.Deref(out.Certificate, "")
	certificateChain := ptr.Deref(out.CertificateChain, "")

	return certificate, certificateChain, nil
}

func (c *acmClient) ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	in := &acm.ListTagsForCertificateInput{
		CertificateArn: new(arn),
	}

	out, err := c.svc.ListTagsForCertificate(ctx, in)
	if err != nil {
		return nil, err
	}

	return out.Tags, nil
}

func (c *acmClient) ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error) {
	// AWS ListCertificates defaults to ISSUED status and RSA_2048 key type only.
	// We explicitly pass all known statuses and key types so nuke can find certificates
	// regardless of their state (e.g. EXPIRED, INACTIVE) or algorithm (e.g. EC_prime256v1).
	// .Values() is called on a zero-value instance — the SDK exposes the full enum list
	// as a method on the type rather than a package-level function.
	// https://docs.aws.amazon.com/acm/latest/APIReference/API_ListCertificates.html
	// https://docs.aws.amazon.com/acm/latest/APIReference/API_Filters.html
	params := &acm.ListCertificatesInput{
		CertificateStatuses: acmtypes.CertificateStatus("").Values(),
		Includes: &acmtypes.Filters{
			KeyTypes: acmtypes.KeyAlgorithm("").Values(),
		},
	}
	out, err := c.svc.ListCertificates(ctx, params)
	if err != nil {
		return nil, err
	}
	return out.CertificateSummaryList, nil
}

func (c *acmClient) ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	out, err := c.svc.ListTagsForCertificate(ctx, &acm.ListTagsForCertificateInput{
		CertificateArn: &arn,
	})
	if err != nil {
		return nil, err
	}
	return out.Tags, nil
}
