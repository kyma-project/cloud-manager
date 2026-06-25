package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"k8s.io/utils/ptr"
)

type Wafv2Client interface {
	CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*wafv2types.WebACL, string, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
}

func NewWafv2Client(svc *wafv2.Client) Wafv2Client {
	return &wafv2Client{
		svc: svc,
	}
}

var _ Wafv2Client = (*wafv2Client)(nil)

type wafv2Client struct {
	svc *wafv2.Client
}

func (c *wafv2Client) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*wafv2types.WebACL, string, error) {
	out, err := c.svc.CreateWebACL(ctx, input)
	if err != nil {
		return nil, "", err
	}

	// Get the full WebACL details
	webACL, lockToken, err := c.GetWebACL(ctx, ptr.Deref(input.Name, ""), ptr.Deref(out.Summary.Id, ""), input.Scope)
	if err != nil {
		return nil, "", err
	}

	return webACL, lockToken, nil
}

func (c *wafv2Client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	in := &wafv2.GetWebACLInput{
		Name:  new(name),
		Id:    new(id),
		Scope: scope,
	}

	out, err := c.svc.GetWebACL(ctx, in)
	if err != nil {
		return nil, "", err
	}

	return out.WebACL, ptr.Deref(out.LockToken, ""), nil
}

func (c *wafv2Client) UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error {
	_, err := c.svc.UpdateWebACL(ctx, input)
	return err
}

func (c *wafv2Client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	in := &wafv2.DeleteWebACLInput{
		Name:      new(name),
		Id:        new(id),
		Scope:     scope,
		LockToken: new(lockToken),
	}

	_, err := c.svc.DeleteWebACL(ctx, in)
	return err
}

func (c *wafv2Client) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	in := &wafv2.ListWebACLsInput{
		Scope: scope,
	}

	out, err := c.svc.ListWebACLs(ctx, in)
	if err != nil {
		return nil, err
	}

	return out.WebACLs, nil
}
