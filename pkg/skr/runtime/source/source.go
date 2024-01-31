package source

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
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

func New(skrManager manager.SkrManager, gvk schema.GroupVersionKind, store clientcache.Store) SkrSource {
	listGvk := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	return &skrSource{
		skrManager:   skrManager,
		scheme:       skrManager.GetScheme(),
		skrApiReader: skrManager.GetAPIReader(),
		objGvk:       gvk,
		listGvk:      listGvk,
		store:        store,
	}
}

type skrSource struct {
	skrManager   manager.SkrManager
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
	logger := s.skrManager.GetLogger().WithValues("GVK", s.objGvk.String())
	logger.Info("Starting SKR Source")
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

	logger.Info(fmt.Sprintf("Loaded list with %d items", len(arr)))

	var events []event.GenericEvent

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

		events = append(events, evt)
		err = s.store.Add(obj)
		if err != nil {
			logger.Error(err, "Error adding object to store")
			return err
		}
	}

	for _, evt := range events {
		handler.Generic(ctx, evt, limitingInterface)
	}

	return nil
}
