package registry

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type SkrRegistry interface {
	Len() int
	Register() Builder
	Builders() []Builder
}

func New(skrScheme *runtime.Scheme) SkrRegistry {
	return &skrRegistry{
		skrScheme: skrScheme,
	}
}

type skrRegistry struct {
	skrScheme *runtime.Scheme
	builders  []Builder
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

func (r *skrRegistry) addBuilder(b *skrBuilder) {
	r.builders = append(r.builders, b)
}
