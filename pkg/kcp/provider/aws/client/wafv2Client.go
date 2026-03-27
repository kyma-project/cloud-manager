package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"k8s.io/utils/ptr"
)

type Wafv2Client interface {
	CreateWebACL(ctx context.Context, name, description string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, tags []wafv2types.Tag) (*wafv2types.WebACL, string, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, lockToken string) error
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

func (c *wafv2Client) CreateWebACL(ctx context.Context, name, description string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, tags []wafv2types.Tag) (*wafv2types.WebACL, string, error) {
	in := &wafv2.CreateWebACLInput{
		Name:             ptr.To(name),
		Scope:            scope,
		DefaultAction:    defaultAction,
		Rules:            rules,
		VisibilityConfig: visibilityConfig,
	}
	if description != "" {
		in.Description = ptr.To(description)
	}
	if len(tags) > 0 {
		in.Tags = tags
	}

	out, err := c.svc.CreateWebACL(ctx, in)
	if err != nil {
		return nil, "", err
	}

	// Get the full WebACL details
	webACL, lockToken, err := c.GetWebACL(ctx, name, ptr.Deref(out.Summary.Id, ""), scope)
	if err != nil {
		return nil, "", err
	}

	return webACL, lockToken, nil
}

func (c *wafv2Client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	in := &wafv2.GetWebACLInput{
		Name:  ptr.To(name),
		Id:    ptr.To(id),
		Scope: scope,
	}

	out, err := c.svc.GetWebACL(ctx, in)
	if err != nil {
		return nil, "", err
	}

	return out.WebACL, ptr.Deref(out.LockToken, ""), nil
}

func (c *wafv2Client) UpdateWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, defaultAction *wafv2types.DefaultAction, rules []wafv2types.Rule, visibilityConfig *wafv2types.VisibilityConfig, lockToken string) error {
	in := &wafv2.UpdateWebACLInput{
		Name:             ptr.To(name),
		Id:               ptr.To(id),
		Scope:            scope,
		DefaultAction:    defaultAction,
		Rules:            rules,
		VisibilityConfig: visibilityConfig,
		LockToken:        ptr.To(lockToken),
	}

	_, err := c.svc.UpdateWebACL(ctx, in)
	return err
}

func (c *wafv2Client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	in := &wafv2.DeleteWebACLInput{
		Name:      ptr.To(name),
		Id:        ptr.To(id),
		Scope:     scope,
		LockToken: ptr.To(lockToken),
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
