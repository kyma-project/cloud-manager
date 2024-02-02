package registry

import (
	skrcache "github.com/kyma-project/cloud-manager/pkg/skr/runtime/cache"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	skrsource "github.com/kyma-project/cloud-manager/pkg/skr/runtime/source"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Watch struct {
	Name         string
	Src          skrsource.SkrSource
	EventHandler handler.EventHandler
	Predicates   []predicate.Predicate
}

type DescriptorList []*Descriptor

func (dl DescriptorList) AllQueuesEmpty() bool {
	for _, desc := range dl {
		for _, watch := range desc.Watches {
			if !watch.Src.IsStarted() {
				return false
			}
			if watch.Src.Queue().Len() > 0 {
				return false
			}
		}
	}
	return true
}

type Descriptor struct {
	Name              string
	GVK               schema.GroupVersionKind
	Watches           []Watch
	ReconcilerFactory reconcile.ReconcilerFactory
}

type SkrRegistry interface {
	Len() int
	GetDescriptors(mngr manager.SkrManager) DescriptorList
	Register() Builder
}

func New(skrScheme *runtime.Scheme) SkrRegistry {
	return &skrRegistry{
		skrScheme: skrScheme,
	}
}

type registryItem struct {
	name              string
	reconcilerFactory reconcile.ReconcilerFactory
	watches           []*registryItemWatch
}

type registryItemWatch struct {
	name         string
	object       client.Object
	gvk          schema.GroupVersionKind
	eventHandler handler.EventHandler
	predicates   []predicate.Predicate
}

type skrRegistry struct {
	skrScheme *runtime.Scheme
	items     []*registryItem
}

func (r *skrRegistry) Len() int {
	return len(r.items)
}

func (r *skrRegistry) GetDescriptors(mngr manager.SkrManager) DescriptorList {
	result := make(DescriptorList, 0, len(r.items))
	cch := mngr.GetCache().(skrcache.SkrCache)
	for _, item := range r.items {
		d := &Descriptor{
			Name:              item.name,
			GVK:               item.watches[0].gvk,
			ReconcilerFactory: item.reconcilerFactory,
		}
		for _, wi := range item.watches {
			src := skrsource.New(mngr, wi.gvk, cch.GetIndexerFor(wi.gvk))
			w := Watch{
				Name:         wi.name,
				Src:          src,
				EventHandler: wi.eventHandler,
				Predicates:   wi.predicates,
			}
			d.Watches = append(d.Watches, w)
		}
		result = append(result, d)
	}
	return result
}

func (r *skrRegistry) Register() Builder {
	return &skrBuilder{registry: r}
}

func (r *skrRegistry) addItem(item *registryItem) {
	r.items = append(r.items, item)
}
