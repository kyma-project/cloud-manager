package testinfra

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clientWithDelegatedReader struct {
	client.Client
	delegatedReader client.Reader
}

func (c *clientWithDelegatedReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return c.delegatedReader.Get(ctx, key, obj, opts...)
}

func (c *clientWithDelegatedReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.delegatedReader.List(ctx, list, opts...)
}
