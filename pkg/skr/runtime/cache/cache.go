package cache

// Based on sigs.k8s.io/controller-runtime@v0.16.3/pkg/cache/internal/cache_reader.go

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	clientcache "k8s.io/client-go/tools/cache"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ SkrCache = &skrCache{}

type SkrCache interface {
	cache.Cache
	GetIndexerFor(gvk schema.GroupVersionKind) clientcache.Indexer
}

func New(
	logger logr.Logger,
	scheme *runtime.Scheme,
	mapper meta.RESTMapper,
) SkrCache {
	return &skrCache{
		logger:   logger,
		scheme:   scheme,
		mapper:   mapper,
		indexers: map[schema.GroupVersionKind]clientcache.Indexer{},
	}
}

// skrCache is the controller-runtime Cache implementation
// that is given to the Client, it only implements the Reader
// since the rest of the Cache methods are not used in the Client
type skrCache struct {
	logger   logr.Logger
	scheme   *runtime.Scheme
	mapper   meta.RESTMapper
	indexers map[schema.GroupVersionKind]clientcache.Indexer
}

func (c *skrCache) GetIndexerFor(gvk schema.GroupVersionKind) clientcache.Indexer {
	i, exists := c.indexers[gvk]
	if !exists {
		i := clientcache.NewIndexer(clientcache.MetaNamespaceKeyFunc, clientcache.Indexers{
			clientcache.NamespaceIndex: clientcache.MetaNamespaceIndexFunc,
		})
		c.indexers[gvk] = i
	}
	return i
}

func (c *skrCache) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
	gvk, err := util.GetObjGvk(c.scheme, out)
	if err != nil {
		return err
	}

	indexer, exists := c.indexers[gvk]
	if !exists {
		return &apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusNotFound,
			Reason: metav1.StatusReasonNotFound,
			Details: &metav1.StatusDetails{
				Group: gvk.Group,
				Kind:  gvk.Kind,
				Name:  key.Name,
			},
			Message: fmt.Sprintf("GVK %s is not cached", gvk),
		}}
	}

	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}
	if mapping.Scope.Name() == apimeta.RESTScopeNameRoot {
		key.Namespace = ""
	}

	storeKey, err := clientcache.MetaNamespaceKeyFunc(out)
	if err != nil {
		return err
	}

	obj, exists, err := indexer.GetByKey(storeKey)
	if err != nil {
		return err
	}
	if !exists {
		return apierrors.NewNotFound(schema.GroupResource{
			Group: gvk.Group,
			// Resource gets set as Kind in the error so this is fine
			Resource: gvk.Kind,
		}, key.Name)
	}

	if _, isObj := obj.(runtime.Object); !isObj {
		// This should never happen
		return fmt.Errorf("cache contained %T, which is not an Object", obj)
	}

	obj = obj.(runtime.Object).DeepCopyObject()

	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("cache had type %s, but %s was asked for", objVal.Type(), outVal.Type())
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	out.GetObjectKind().SetGroupVersionKind(gvk)

	return nil
}

func (c *skrCache) List(ctx context.Context, out client.ObjectList, opts ...client.ListOption) error {
	listGvk, err := util.GetObjGvk(c.scheme, out)
	if err != nil {
		return err
	}
	gvk, err := util.GetObjGvkFromListGkv(c.scheme, listGvk)
	if err != nil {
		return err
	}

	indexer, exists := c.indexers[gvk]
	if !exists {
		return &apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusNotFound,
			Reason: metav1.StatusReasonNotFound,
			Details: &metav1.StatusDetails{
				Group: gvk.Group,
				Kind:  gvk.Kind,
			},
			Message: fmt.Sprintf("GVK %s is not cached", gvk),
		}}
	}

	var objs []interface{}

	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	if listOpts.Continue != "" {
		return fmt.Errorf("continue list option is not supported by the cache")
	}

	switch {
	case listOpts.FieldSelector != nil:
		field, val, requiresExact := requiresExactMatch(listOpts.FieldSelector)
		if !requiresExact {
			return fmt.Errorf("non-exact field matches are not supported by the cache")
		}
		// list all objects by the field selector. If this is namespaced and we have one, ask for the
		// namespaced index key. Otherwise, ask for the non-namespaced variant by using the fake "all namespaces"
		// namespace.
		objs, err = indexer.ByIndex(fieldIndexName(field), keyToNamespacedKey(listOpts.Namespace, val))
	case listOpts.Namespace != "":
		objs, err = indexer.ByIndex(clientcache.NamespaceIndex, listOpts.Namespace)
	default:
		objs = indexer.List()
	}
	if err != nil {
		return err
	}
	var labelSel labels.Selector
	if listOpts.LabelSelector != nil {
		labelSel = listOpts.LabelSelector
	}

	limitSet := listOpts.Limit > 0

	runtimeObjs := make([]runtime.Object, 0, len(objs))
	for _, item := range objs {
		// if the Limit option is set and the number of items
		// listed exceeds this limit, then stop reading.
		if limitSet && int64(len(runtimeObjs)) >= listOpts.Limit {
			break
		}
		obj, isObj := item.(runtime.Object)
		if !isObj {
			return fmt.Errorf("cache contained %T, which is not an Object", item)
		}
		metaAcc, err := apimeta.Accessor(obj)
		if err != nil {
			return err
		}
		if labelSel != nil {
			lbls := labels.Set(metaAcc.GetLabels())
			if !labelSel.Matches(lbls) {
				continue
			}
		}

		var outObj runtime.Object
		if listOpts.UnsafeDisableDeepCopy != nil && *listOpts.UnsafeDisableDeepCopy {
			// skip deep copy which might be unsafe
			// you must DeepCopy any object before mutating it outside
			outObj = obj
		} else {
			outObj = obj.DeepCopyObject()
			outObj.GetObjectKind().SetGroupVersionKind(gvk)
		}
		runtimeObjs = append(runtimeObjs, outObj)
	}
	return apimeta.SetList(out, runtimeObjs)
}

func (c *skrCache) GetInformer(ctx context.Context, obj client.Object, opts ...cache.InformerGetOption) (cache.Informer, error) {
	//TODO implement me
	panic("implement me")
}

func (c *skrCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...cache.InformerGetOption) (cache.Informer, error) {
	//TODO implement me
	panic("implement me")
}

func (c *skrCache) Start(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (c *skrCache) WaitForCacheSync(ctx context.Context) bool {
	//TODO implement me
	panic("implement me")
}

func (c *skrCache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	//TODO implement me
	panic("implement me")
}

func requiresExactMatch(sel fields.Selector) (field, val string, required bool) {
	reqs := sel.Requirements()
	if len(reqs) != 1 {
		return "", "", false
	}
	req := reqs[0]
	if req.Operator != selection.Equals && req.Operator != selection.DoubleEquals {
		return "", "", false
	}
	return req.Field, req.Value, true
}

func fieldIndexName(field string) string {
	return "field:" + field
}

// allNamespacesNamespace is used as the "namespace" when we want to list across all namespaces.
const allNamespacesNamespace = "__all_namespaces"

// KeyToNamespacedKey prefixes the given index key with a namespace
// for use in field selector indexes.
func keyToNamespacedKey(ns string, baseKey string) string {
	if ns != "" {
		return ns + "/" + baseKey
	}
	return allNamespacesNamespace + "/" + baseKey
}
