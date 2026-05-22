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
		CertificateArn: ptr.To(arn),
	}

	out, err := c.svc.DescribeCertificate(ctx, in)
	if err != nil {
		return nil, err
	}

	return out.Certificate, nil
}

func (c *acmClient) DeleteCertificate(ctx context.Context, arn string) error {
	in := &acm.DeleteCertificateInput{
		CertificateArn: ptr.To(arn),
	}

	_, err := c.svc.DeleteCertificate(ctx, in)
	return err
}

func (c *acmClient) GetCertificate(ctx context.Context, arn string) (string, string, error) {
	in := &acm.GetCertificateInput{
		CertificateArn: ptr.To(arn),
	}

	out, err := c.svc.GetCertificate(ctx, in)
	if err != nil {
		return "", "", err
	}

	certificate := ptr.Deref(out.Certificate, "")
	certificateChain := ptr.Deref(out.CertificateChain, "")

	return certificate, certificateChain, nil
}
