package registry

import (
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"strings"
)

var _ Builder = &skrBuilder{}

type Builder interface {
	WithFactory(f ReconcilerFactory) Builder

	Named(name string) Builder
	For(object client.Object, opts ...ForOption) Builder
	Watches(object client.Object, eventHandler handler.EventHandler, opts ...WatchesOption) Builder
	Complete() error
}

type skrBuilder struct {
	registry     *skrRegistry
	factory      ReconcilerFactory
	name         string
	forInput     *ForInput
	watchesInput []*WatchesInput
}

func (b *skrBuilder) Named(name string) Builder {
	b.name = name
	return b
}

func (b *skrBuilder) WithFactory(f ReconcilerFactory) Builder {
	b.factory = f
	return b
}

func (b *skrBuilder) For(object client.Object, opts ...ForOption) Builder {
	in := &ForInput{object: object}
	for _, opt := range opts {
		opt.ApplyToFor(in)
	}
	b.forInput = in
	return b
}

func (b *skrBuilder) Watches(object client.Object, eventHandler handler.EventHandler, opts ...WatchesOption) Builder {
	in := &WatchesInput{object: object}
	for _, opt := range opts {
		opt.ApplyToWatches(in)
	}
	b.watchesInput = append(b.watchesInput, in)
	return b
}

func (b *skrBuilder) Complete() error {
	if b.factory == nil {
		return errors.New("the WithFactory(...) method must be called to define reconciler factory")
	}
	if b.forInput == nil {
		return errors.New("the For(...) method must be called to define reconciled object")
	}

	gvk, err := util.GetObjGvk(b.registry.skrScheme, b.forInput.object)
	if err != nil {
		return err
	}

	name := b.name
	if len(name) == 0 {
		name = strings.ToLower(gvk.Kind)
	}
	item := &registryItem{
		name:              name,
		reconcilerFactory: b.factory,
	}

	gvk, err = util.GetObjGvk(b.registry.skrScheme, b.forInput.object)
	if err != nil {
		return err
	}
	hdler := &handler.EnqueueRequestForObject{}
	item.watches = append(item.watches, &registryItemWatch{
		name:         fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind),
		object:       b.forInput.object,
		gvk:          gvk,
		eventHandler: hdler,
		predicates:   nil,
	})

	for _, w := range b.watchesInput {
		gvk, err := util.GetObjGvk(b.registry.skrScheme, w.object)
		if err != nil {
			return err
		}
		wi := &registryItemWatch{
			name:         gvk.String(),
			object:       w.object,
			gvk:          gvk,
			eventHandler: w.eventHandler,
			predicates:   w.predicates,
		}

		item.watches = append(item.watches, wi)
	}

	b.registry.addItem(item)

	return nil
}
