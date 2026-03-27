package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	CreateWebACL(ctx context.Context, name, description string, scope types.Scope, defaultAction *types.DefaultAction, rules []types.Rule, visibilityConfig *types.VisibilityConfig, tags []types.Tag) (*types.WebACL, string, error)
	GetWebACL(ctx context.Context, name, id string, scope types.Scope) (*types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, name, id string, scope types.Scope, defaultAction *types.DefaultAction, rules []types.Rule, visibilityConfig *types.VisibilityConfig, lockToken string) error
	DeleteWebACL(ctx context.Context, name, id string, scope types.Scope, lockToken string) error
	ListWebACLs(ctx context.Context, scope types.Scope) ([]types.WebACLSummary, error)
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(awsclient.NewWafv2Client(wafv2.NewFromConfig(cfg))), nil
	}
}

func newClient(wafv2Client awsclient.Wafv2Client) Client { return &client{Wafv2Client: wafv2Client} }

var _ Client = (*client)(nil)

type client struct {
	awsclient.Wafv2Client
}
