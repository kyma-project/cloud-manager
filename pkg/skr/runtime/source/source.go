package source

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	ObjGVK() schema.GroupVersionKind
	IsStarted() bool
	Queue() workqueue.RateLimitingInterface
	LoadAll(ctx context.Context) error
	LoadOne(ctx context.Context, key types.NamespacedName) error
}

func New(skrManager manager.SkrManager, gvk schema.GroupVersionKind, store clientcache.Store) SkrSource {
	listGvk := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	return &skrSource{
		skrManager:   skrManager,
		skrScheme:    skrManager.GetScheme(),
		skrApiReader: skrManager.GetAPIReader(),
		objGvk:       gvk,
		listGvk:      listGvk,
		store:        store,
	}
}

type skrSource struct {
	skrManager   manager.SkrManager
	skrScheme    *runtime.Scheme
	skrApiReader client.Reader
	objGvk       schema.GroupVersionKind
	listGvk      schema.GroupVersionKind
	store        clientcache.Store

	started    bool
	queue      workqueue.RateLimitingInterface
	predicates []predicate.Predicate
	handler    handler.EventHandler
	logger     logr.Logger
}

func (s *skrSource) ObjGVK() schema.GroupVersionKind {
	return s.objGvk
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

func (s *skrSource) Start(ctx context.Context, handler handler.EventHandler, limitingInterface workqueue.RateLimitingInterface, predicates ...predicate.Predicate) error {
	s.logger = s.skrManager.GetLogger().WithValues("GVK", s.objGvk.String())
	s.logger.Info("Starting SKR Source")
	s.started = true
	s.queue = limitingInterface
	s.handler = handler
	s.predicates = predicates

	return s.LoadAll(ctx)
}

func (s *skrSource) LoadAll(ctx context.Context) error {
	obj, err := s.skrScheme.New(s.listGvk)
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

	s.logger.Info(fmt.Sprintf("Loaded list with %d items", len(arr)))

	err = s.processObjects(ctx, arr, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *skrSource) LoadOne(ctx context.Context, key types.NamespacedName) error {
	o, err := s.skrScheme.New(s.objGvk)
	if err != nil {
		return fmt.Errorf("error instantiating new object for %s: %w", s.objGvk, err)
	}
	obj, ok := o.(client.Object)
	if !ok {
		return fmt.Errorf("the GVK %s is not a client.Object", s.objGvk)
	}

	err = s.skrApiReader.Get(ctx, key, obj)
	if err != nil {
		return fmt.Errorf("error reloading one %s with key %s: %w", s.objGvk, key, err)
	}

	err = s.processObjects(ctx, []runtime.Object{obj}, false)
	if err != nil {
		return err
	}

	return nil
}

func (s *skrSource) processObjects(ctx context.Context, objects []runtime.Object, removeFromStore bool) error {

	var events []event.GenericEvent

	// filter loaded objects with predicates
itemLoop:
	for _, item := range objects {
		obj, ok := item.(client.Object)
		if !ok {
			continue
		}
		evt := event.GenericEvent{Object: obj}

		for _, p := range s.predicates {
			if !p.Generic(evt) {
				continue itemLoop
			}
		}

		events = append(events, evt)
	}

	// ensure new objects are added to store
	for _, evt := range events {
		err := s.store.Add(evt.Object)
		if err != nil {
			s.logger.Error(err, "Error adding object to store")
			return err
		}
	}
	// remove object from store that are not loaded now
	if removeFromStore {
		loadedObjects := make(map[string]struct{}, len(events))
		for _, evt := range events {
			key, _ := clientcache.MetaNamespaceKeyFunc(evt.Object)
			loadedObjects[key] = struct{}{}
		}

		for _, xObj := range s.store.List() {
			obj, ok := xObj.(client.Object)
			if !ok {
				continue
			}
			key, err := clientcache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				continue
			}
			_, isLoaded := loadedObjects[key]
			if !isLoaded {
				_ = s.store.Delete(xObj)
			}
		}
	}

	// process loaded objects
	for _, evt := range events {
		s.handler.Generic(ctx, evt, s.queue)
	}

	return nil
}
