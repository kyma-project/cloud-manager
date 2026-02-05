package client

import (
	"context"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/googleapis/gax-go/v2"
)

type FilestoreClient interface {
	GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, opts ...gax.CallOption) (*filestorepb.Instance, error)
	CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, opts ...gax.CallOption)
}

var _ FilestoreClient = (*filestoreClient)(nil)

type filestoreClient struct {
	inner *filestore.CloudFilestoreManagerClient
}

func (c *filestoreClient) GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, opts ...gax.CallOption) (*filestorepb.Instance, error) {
	return c.inner.GetInstance(ctx, req, opts...)
}

func (c *filestoreClient) CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, opts ...gax.CallOption) {
	op, err := c.inner.CreateInstance(ctx, req, opts...)
}
