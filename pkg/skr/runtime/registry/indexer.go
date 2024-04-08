package registry

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SkrIndexer interface {
	Obj() client.Object
	Field() string
	IndexField(ctx context.Context, indexer client.FieldIndexer) error
}

type skrIndexer struct {
	obj          client.Object
	field        string
	extractValue client.IndexerFunc
}

func (si *skrIndexer) Obj() client.Object {
	return si.obj
}

func (si *skrIndexer) Field() string {
	return si.field
}

func (si *skrIndexer) IndexField(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, si.obj, si.field, si.extractValue)
}
