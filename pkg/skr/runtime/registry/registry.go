package registry

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SkrRegistry interface {
	Len() int
	Register() Builder
	Builders() []Builder

	IndexField(obj client.Object, field string, extractValue client.IndexerFunc)
	Indexers() []SkrIndexer
}

func New(skrScheme *runtime.Scheme) SkrRegistry {
	return &skrRegistry{
		skrScheme: skrScheme,
	}
}

type skrRegistry struct {
	skrScheme *runtime.Scheme
	builders  []Builder
	indexers  []SkrIndexer
}

func (r *skrRegistry) Builders() []Builder {
	return r.builders
}

func (r *skrRegistry) Len() int {
	return len(r.builders)
}

func (r *skrRegistry) Register() Builder {
	return &skrBuilder{registry: r}
}

func (r *skrRegistry) IndexField(obj client.Object, field string, extractValue client.IndexerFunc) {
	r.indexers = append(r.indexers, &skrIndexer{
		obj:          obj,
		field:        field,
		extractValue: extractValue,
	})
}

func (r *skrRegistry) Indexers() []SkrIndexer {
	return r.indexers
}

func (r *skrRegistry) addBuilder(b *skrBuilder) {
	r.builders = append(r.builders, b)
}
