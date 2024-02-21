package registry

import (
	"errors"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ Builder = &skrBuilder{}

type Builder interface {
	WithFactory(f reconcile.ReconcilerFactory) Builder

	GetForObj() client.Object

	For(object client.Object, opts ...builder.ForOption) Builder
	Owns(object client.Object, opts ...builder.OwnsOption) Builder
	Watches(object client.Object, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder
	WatchesMetadata(object client.Object, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder
	WatchesRawSource(src source.Source, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder
	WithEventFilter(p predicate.Predicate) Builder
	WithOptions(options controller.Options) Builder
	WithLogConstructor(logConstructor func(*ctrlreconcile.Request) logr.Logger) Builder
	Named(name string) Builder

	Complete() error

	SetupWithManager(mngr manager.Manager, args reconcile.ReconcilerArguments) error
}

type applyBuildItem func(cb *builder.Builder)

type skrBuilder struct {
	registry     *skrRegistry
	factory      reconcile.ReconcilerFactory
	items        []applyBuildItem
	forCallCount int
	forObj       client.Object
}

func (b *skrBuilder) add(i applyBuildItem) Builder {
	b.items = append(b.items, i)
	return b
}

func (b *skrBuilder) WithFactory(f reconcile.ReconcilerFactory) Builder {
	b.factory = f
	return b
}

func (b *skrBuilder) GetForObj() client.Object {
	return b.forObj
}

func (b *skrBuilder) For(object client.Object, opts ...builder.ForOption) Builder {
	b.forCallCount++
	b.forObj = object
	return b.add(func(cb *builder.Builder) {
		cb.For(object, opts...)
	})
}

func (b *skrBuilder) Owns(object client.Object, opts ...builder.OwnsOption) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.Owns(object, opts...)
	})
}

func (b *skrBuilder) Watches(object client.Object, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.Watches(object, eventHandler, opts...)
	})
}

func (b *skrBuilder) WatchesMetadata(object client.Object, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.WatchesMetadata(object, eventHandler, opts...)
	})
}

func (b *skrBuilder) WatchesRawSource(src source.Source, eventHandler handler.EventHandler, opts ...builder.WatchesOption) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.WatchesRawSource(src, eventHandler, opts...)
	})
}

func (b *skrBuilder) WithEventFilter(p predicate.Predicate) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.WithEventFilter(p)
	})
}

func (b *skrBuilder) WithOptions(options controller.Options) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.WithOptions(options)
	})
}

func (b *skrBuilder) WithLogConstructor(logConstructor func(*ctrlreconcile.Request) logr.Logger) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.WithLogConstructor(logConstructor)
	})
}

func (b *skrBuilder) Named(name string) Builder {
	return b.add(func(cb *builder.Builder) {
		cb.Named(name)
	})
}

func (b *skrBuilder) Complete() error {
	if b.factory == nil {
		return errors.New("method WithFactory() must be called")
	}
	if b.forCallCount > 1 {
		return errors.New("method For() should only be called once, could not assign multiple objects for reconciliation")
	}
	b.registry.addBuilder(b)
	return nil
}

func (b *skrBuilder) SetupWithManager(mngr manager.Manager, args reconcile.ReconcilerArguments) error {
	r := b.factory.New(args)
	cb := ctrl.NewControllerManagedBy(mngr)
	for _, i := range b.items {
		i(cb)
	}
	return cb.Complete(r)
}
