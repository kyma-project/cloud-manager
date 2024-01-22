package source

import (
	"context"
	"fmt"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	ctrlsource "sigs.k8s.io/controller-runtime/pkg/source"
)

var _ SkrSource = &skrSource{}

type SkrSource interface {
	ctrlsource.Source
	fmt.Stringer
	IsStarted() bool
	Queue() workqueue.RateLimitingInterface
}

func New(scheme *runtime.Scheme, skrApiReader client.Reader, gvk schema.GroupVersionKind, store clientcache.Store) SkrSource {
	listGvk := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	return &skrSource{
		scheme:       scheme,
		skrApiReader: skrApiReader,
		objGvk:       gvk,
		listGvk:      listGvk,
		store:        store,
	}
}

type skrSource struct {
	scheme       *runtime.Scheme
	skrApiReader client.Reader
	objGvk       schema.GroupVersionKind
	listGvk      schema.GroupVersionKind
	store        clientcache.Store

	started bool
	queue   workqueue.RateLimitingInterface
}

func (s *skrSource) String() string {
	return fmt.Sprintf("skr kind: %s", s.objGvk.Kind)
}

func (s *skrSource) IsStarted() bool {
	return s.started
}

func (s *skrSource) Queue() workqueue.RateLimitingInterface {
	return s.queue
}

func (s *skrSource) Start(ctx context.Context, handler handler.EventHandler, limitingInterface workqueue.RateLimitingInterface, predicate ...predicate.Predicate) error {
	s.started = true
	s.queue = limitingInterface

	obj, err := s.scheme.New(s.listGvk)
	if err != nil {
		return fmt.Errorf("error instantiating new list object for %s: %w", s.listGvk, err)
	}
	list, ok := obj.(client.ObjectList)
	if !ok {
		return fmt.Errorf("the GVK %s is not a client.ObjectList", s.listGvk)
	}

	err = s.skrApiReader.List(ctx, list)
	if err != nil {
		return err
	}
	arr, err := apimeta.ExtractList(list)
	if err != nil {
		return err
	}
itemLoop:
	for _, item := range arr {
		obj, ok := item.(client.Object)
		if !ok {
			continue
		}
		evt := event.GenericEvent{Object: obj}

		for _, p := range predicate {
			if !p.Generic(evt) {
				continue itemLoop
			}
		}

		handler.Generic(ctx, evt, limitingInterface)
	}

	return nil
}
