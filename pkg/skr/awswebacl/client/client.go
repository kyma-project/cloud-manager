package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	CreateWebACL(ctx context.Context, name, description string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, tags []wafv2types.Tag) (*wafv2types.WebACL, string, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, lockToken string) error
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
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

func newClient(wafv2Client awsclient.Wafv2Client) Client { return &client{wafv2Client: wafv2Client} }

var _ Client = (*client)(nil)

type client struct {
	wafv2Client awsclient.Wafv2Client
}

func (c *client) CreateWebACL(ctx context.Context, name, description string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, tags []wafv2types.Tag) (*wafv2types.WebACL, string, error) {
	return c.wafv2Client.CreateWebACL(ctx, name, description, scope, defaultAction, rules, visibilityConfig, tags)
}

func (c *client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	return c.wafv2Client.GetWebACL(ctx, name, id, scope)
}

func (c *client) UpdateWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, lockToken string) error {
	return c.wafv2Client.UpdateWebACL(ctx, name, id, scope, defaultAction, rules, visibilityConfig, lockToken)
}

func (c *client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	return c.wafv2Client.DeleteWebACL(ctx, name, id, scope, lockToken)
}

func (c *client) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	return c.wafv2Client.ListWebACLs(ctx, scope)
}
