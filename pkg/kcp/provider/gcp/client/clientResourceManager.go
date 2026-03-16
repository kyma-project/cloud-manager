package client

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
)

type ResourceManagerClient interface {
	// Keys

	GetTagKey(ctx context.Context, req *resourcemanagerpb.GetTagKeyRequest, opts ...gax.CallOption) (*resourcemanagerpb.TagKey, error)
	CreateTagKey(ctx context.Context, req *resourcemanagerpb.CreateTagKeyRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagKey], error)
	DeleteTagKey(ctx context.Context, req *resourcemanagerpb.DeleteTagKeyRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagKey], error)
	ListTagKeys(ctx context.Context, req *resourcemanagerpb.ListTagKeysRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagKey]

	// Values

	GetTagValue(ctx context.Context, req *resourcemanagerpb.GetTagValueRequest, opts ...gax.CallOption) (*resourcemanagerpb.TagValue, error)
	CreateTagValue(ctx context.Context, req *resourcemanagerpb.CreateTagValueRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagValue], error)
	DeleteTagValue(ctx context.Context, req *resourcemanagerpb.DeleteTagValueRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagValue], error)
	ListTagValues(ctx context.Context, req *resourcemanagerpb.ListTagValuesRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagValue]

	// Bindings

	CreateTagBinding(ctx context.Context, req *resourcemanagerpb.CreateTagBindingRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagBinding], error)
	DeleteTagBinding(ctx context.Context, req *resourcemanagerpb.DeleteTagBindingRequest, opts ...gax.CallOption) (VoidOperation, error)
	ListTagBindings(ctx context.Context, req *resourcemanagerpb.ListTagBindingsRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagBinding]
	ListEffectiveTags(ctx context.Context, req *resourcemanagerpb.ListEffectiveTagsRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.EffectiveTag]
}

type resourceManagerClient struct {
	keysClient     *resourcemanager.TagKeysClient
	valuesClient   *resourcemanager.TagValuesClient
	bindingsClient *resourcemanager.TagBindingsClient
}

var _ ResourceManagerClient = (*resourceManagerClient)(nil)

// Keys ===========================================================================

func (c *resourceManagerClient) GetTagKey(ctx context.Context, req *resourcemanagerpb.GetTagKeyRequest, opts ...gax.CallOption) (*resourcemanagerpb.TagKey, error) {
	return c.keysClient.GetTagKey(ctx, req, opts...)
}

func (c *resourceManagerClient) CreateTagKey(ctx context.Context, req *resourcemanagerpb.CreateTagKeyRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagKey], error) {
	return c.keysClient.CreateTagKey(ctx, req, opts...)
}

func (c *resourceManagerClient) DeleteTagKey(ctx context.Context, req *resourcemanagerpb.DeleteTagKeyRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagKey], error) {
	return c.keysClient.DeleteTagKey(ctx, req, opts...)
}

func (c *resourceManagerClient) ListTagKeys(ctx context.Context, req *resourcemanagerpb.ListTagKeysRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagKey] {
	return c.keysClient.ListTagKeys(ctx, req, opts...)
}

// Values ===========================================================================

func (c *resourceManagerClient) GetTagValue(ctx context.Context, req *resourcemanagerpb.GetTagValueRequest, opts ...gax.CallOption) (*resourcemanagerpb.TagValue, error) {
	return c.valuesClient.GetTagValue(ctx, req, opts...)
}

func (c *resourceManagerClient) CreateTagValue(ctx context.Context, req *resourcemanagerpb.CreateTagValueRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagValue], error) {
	return c.valuesClient.CreateTagValue(ctx, req, opts...)
}

func (c *resourceManagerClient) DeleteTagValue(ctx context.Context, req *resourcemanagerpb.DeleteTagValueRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagValue], error) {
	return c.valuesClient.DeleteTagValue(ctx, req, opts...)
}

func (c *resourceManagerClient) ListTagValues(ctx context.Context, req *resourcemanagerpb.ListTagValuesRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagValue] {
	return c.valuesClient.ListTagValues(ctx, req, opts...)
}

// Bindings ===========================================================================

func (c *resourceManagerClient) CreateTagBinding(ctx context.Context, req *resourcemanagerpb.CreateTagBindingRequest, opts ...gax.CallOption) (ResultOperation[*resourcemanagerpb.TagBinding], error) {
	return c.bindingsClient.CreateTagBinding(ctx, req, opts...)
}

func (c *resourceManagerClient) DeleteTagBinding(ctx context.Context, req *resourcemanagerpb.DeleteTagBindingRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.bindingsClient.DeleteTagBinding(ctx, req, opts...)
}

func (c *resourceManagerClient) ListTagBindings(ctx context.Context, req *resourcemanagerpb.ListTagBindingsRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.TagBinding] {
	return c.bindingsClient.ListTagBindings(ctx, req, opts...)
}

func (c *resourceManagerClient) ListEffectiveTags(ctx context.Context, req *resourcemanagerpb.ListEffectiveTagsRequest, opts ...gax.CallOption) Iterator[*resourcemanagerpb.EffectiveTag] {
	return c.bindingsClient.ListEffectiveTags(ctx, req, opts...)
}
