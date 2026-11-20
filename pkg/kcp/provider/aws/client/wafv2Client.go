package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"k8s.io/utils/ptr"
)

type Wafv2Client interface {
	CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) error
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
	ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error)
	PutLoggingConfiguration(ctx context.Context, webAclArn, logGroupArn string) error
	DeleteLoggingConfiguration(ctx context.Context, webAclArn string) error
	GetLoggingConfiguration(ctx context.Context, webAclArn string) (*wafv2types.LoggingConfiguration, error)
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

func (c *wafv2Client) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) error {
	_, err := c.svc.CreateWebACL(ctx, input)
	return err
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
	var result []wafv2types.WebACLSummary
	in := &wafv2.ListWebACLsInput{
		Scope: scope,
	}

	for {
		out, err := c.svc.ListWebACLs(ctx, in)
		if err != nil {
			return nil, err
		}
		result = append(result, out.WebACLs...)
		if out.NextMarker == nil {
			break
		}
		in.NextMarker = out.NextMarker
	}

	return result, nil
}

func (c *wafv2Client) ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error) {
	out, err := c.svc.ListTagsForResource(ctx, &wafv2.ListTagsForResourceInput{
		ResourceARN: &resourceArn,
	})
	if err != nil {
		return nil, err
	}
	if out.TagInfoForResource == nil {
		return nil, nil
	}
	return out.TagInfoForResource.TagList, nil
}

func (c *wafv2Client) PutLoggingConfiguration(ctx context.Context, webAclArn, logGroupArn string) error {
	_, err := c.svc.PutLoggingConfiguration(ctx, &wafv2.PutLoggingConfigurationInput{
		LoggingConfiguration: &wafv2types.LoggingConfiguration{
			ResourceArn:           &webAclArn,
			LogDestinationConfigs: []string{logGroupArn},
		},
	})
	return err
}

func (c *wafv2Client) DeleteLoggingConfiguration(ctx context.Context, webAclArn string) error {
	_, err := c.svc.DeleteLoggingConfiguration(ctx, &wafv2.DeleteLoggingConfigurationInput{
		ResourceArn: &webAclArn,
	})
	return err
}

func (c *wafv2Client) GetLoggingConfiguration(ctx context.Context, webAclArn string) (*wafv2types.LoggingConfiguration, error) {
	out, err := c.svc.GetLoggingConfiguration(ctx, &wafv2.GetLoggingConfigurationInput{
		ResourceArn: &webAclArn,
	})
	if err != nil {
		return nil, err
	}
	return out.LoggingConfiguration, nil
}
