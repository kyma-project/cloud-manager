package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error)
	DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error)
	GetCertificate(ctx context.Context, arn string) (certificate string, certificateChain string, err error)
	DeleteCertificate(ctx context.Context, arn string) error
	SearchCertificates(ctx context.Context, input *acm.SearchCertificatesInput) ([]acmtypes.CertificateSearchResult, error)
	ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error)
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(
			awsclient.NewAcmClient(acm.NewFromConfig(cfg)),
		), nil
	}
}

func newClient(acmClient awsclient.AcmClient) Client {
	return &client{acmClient: acmClient}
}

var _ Client = (*client)(nil)

type client struct {
	acmClient awsclient.AcmClient
}

func (c *client) ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error) {
	return c.acmClient.ImportCertificate(ctx, input)
}

func (c *client) DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error) {
	return c.acmClient.DescribeCertificate(ctx, arn)
}

func (c *client) DeleteCertificate(ctx context.Context, arn string) error {
	return c.acmClient.DeleteCertificate(ctx, arn)
}

func (c *client) GetCertificate(ctx context.Context, arn string) (string, string, error) {
	return c.acmClient.GetCertificate(ctx, arn)
}

func (c *client) SearchCertificates(ctx context.Context, input *acm.SearchCertificatesInput) ([]acmtypes.CertificateSearchResult, error) {
	return c.acmClient.SearchCertificates(ctx, input)
}

func (c *client) ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	return c.acmClient.ListTagsForCertificate(ctx, arn)
}
