package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"k8s.io/utils/ptr"
)

type AcmClient interface {
	ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error)
	DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error)
	GetCertificate(ctx context.Context, arn string) (certificate string, certificateChain string, err error)
	DeleteCertificate(ctx context.Context, arn string) error
	SearchCertificates(ctx context.Context, input *acm.SearchCertificatesInput) ([]acmtypes.CertificateSearchResult, error)
	ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error)
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

func (c *acmClient) SearchCertificates(ctx context.Context, input *acm.SearchCertificatesInput) ([]acmtypes.CertificateSearchResult, error) {
	result, err := c.svc.SearchCertificates(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("SearchCertificates failed: %w", err)
	}

	return result.Results, nil
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
