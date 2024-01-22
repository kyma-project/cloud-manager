package registry

import (
	skrcache "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/cache"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/manager"
	skrsource "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/source"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ReconcilerFactory interface {
	New(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster) reconcile.Reconciler
}

type Watch struct {
	Name         string
	Src          source.Source
	EventHandler handler.EventHandler
	Predicates   []predicate.Predicate
}

type DescriptorList []Descriptor

func (dl DescriptorList) AllQueuesEmpty() bool {
	for _, desc := range dl {
		for _, watch := range desc.Watches {
			src, ok := watch.Src.(skrsource.SkrSource)
			if !ok {
				continue
			}
			if !src.IsStarted() {
				return false
			}
			if src.Queue().Len() > 0 {
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
	ReconcilerFactory ReconcilerFactory
}

type SkrRegistry interface {
	GetDescriptors(mngr manager.SkrManager) DescriptorList
	Register() Builder
}

func New(scheme *runtime.Scheme) SkrRegistry {
	return &skrRegistry{
		scheme: scheme,
	}
}

type registryItem struct {
	name              string
	reconcilerFactory ReconcilerFactory
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
	scheme *runtime.Scheme
	items  []*registryItem
}

func (r *skrRegistry) GetDescriptors(mngr manager.SkrManager) DescriptorList {
	result := make(DescriptorList, 0, len(r.items))
	cch := mngr.GetCache().(skrcache.SkrCache)
	for _, item := range r.items {
		d := Descriptor{
			Name:              item.name,
			GVK:               item.watches[0].gvk,
			ReconcilerFactory: item.reconcilerFactory,
		}
		for _, wi := range item.watches {
			src := skrsource.New(mngr.GetScheme(), mngr.GetAPIReader(), wi.gvk, cch.GetIndexerFor(wi.gvk))
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
