package e2e

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Cluster interface {
	cluster.Cluster

	Alias() string

	// AddResources declares k8s objects that are watched and can be got from the cache
	AddResources(ctx context.Context, arr ...*ResourceDeclaration) error

	// Has returns true if resource by given alias is declared
	Has(alias string) bool

	// Get returns resource declarede with the given alias from the cache. If resource does not exist, nil object
	// and not error are returned. Also, if kind is not registered, nil resource and no error are returned.
	// Error is returned only if it's some other than NotFound and NoMatch
	Get(ctx context.Context, alias string) (*unstructured.Unstructured, error)

	// EvaluationContext returns map of all declared resources
	EvaluationContext(ctx context.Context) (map[string]interface{}, error)
}

func NewCluster(alias string, clstr cluster.Cluster) Cluster {
	return &defaultCluster{
		Cluster:   clstr,
		alias:     alias,
		resources: make(map[string]*ResourceInfo),
		sources:   make(map[schema.GroupVersionKind]source.SyncingSource),
	}
}

type defaultCluster struct {
	cluster.Cluster

	alias string

	runCtx context.Context

	resources map[string]*ResourceInfo
	sources   map[schema.GroupVersionKind]source.SyncingSource
}

func (c *defaultCluster) Start(ctx context.Context) error {
	c.runCtx = ctx
	return c.Cluster.Start(ctx)
}

func (c *defaultCluster) Alias() string {
	return c.alias
}

func (c *defaultCluster) AddResources(ctx context.Context, arr ...*ResourceDeclaration) error {
	if c.runCtx == nil {
		return fmt.Errorf("runCtx is nil, check if the cluster is started: %w", common.ErrLogical)
	}
	addedSources := map[schema.GroupVersionKind]source.SyncingSource{}

	for _, decl := range arr {
		if _, ok := c.resources[decl.Alias]; ok {
			return fmt.Errorf("alias %s already declared", decl.Alias)
		}

		gv, err := schema.ParseGroupVersion(decl.ApiVersion)
		if err != nil {
			return fmt.Errorf("failed to parse GroupVersion %s: %w", decl.ApiVersion, err)
		}
		gvk := gv.WithKind(decl.Kind)

		ri := &ResourceInfo{
			ResourceDeclaration: *decl,
			Evaluated:           false,
			GVK:                 gvk,
		}
		c.resources[decl.Alias] = ri

		src, sourceExists := c.sources[gvk]
		if sourceExists {
			ri.Source = src
			continue
		}

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)

		src = source.Kind(c.Cluster.GetCache(), u, handler.TypedFuncs[*unstructured.Unstructured, reconcile.Request]{})
		err = src.Start(c.runCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to start source for alias %q GVK %s: %w", decl.Alias, gvk.String(), err)
		}
		addedSources[gvk] = src

	}

	for gvk, src := range addedSources {
		err := src.WaitForSync(c.runCtx)
		if err != nil {
			return fmt.Errorf("failed to wait for source sync for GVK %s: %w", gvk.String(), err)
		}
	}
	if len(addedSources) > 0 {
		ok := c.Cluster.GetCache().WaitForCacheSync(ctx)
		if !ok {
			return fmt.Errorf("failed to wait for cache sync after adding resources")
		}

	}

	return nil
}

func (c *defaultCluster) Has(alias string) bool {
	_, ok := c.resources[alias]
	return ok
}

func (c *defaultCluster) Get(ctx context.Context, alias string) (*unstructured.Unstructured, error) {
	ri, ok := c.resources[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not found", alias)
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(ri.GVK)

	err := c.GetClient().Get(ctx, types.NamespacedName{
		Namespace: ri.Namespace,
		Name:      ri.Name,
	}, u)
	if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error loading resource %q", alias)
	}
	return u, nil
}

func (c *defaultCluster) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(c.resources))
	for _, ri := range c.resources {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(ri.GVK)

		err := c.Cluster.GetClient().Get(ctx, types.NamespacedName{
			Namespace: ri.Namespace,
			Name:      ri.Name,
		}, u)
		if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
			return nil, fmt.Errorf("failed to load object for alias %q: %w", ri.Alias, err)
		}
		if err != nil {
			result[ri.Alias] = nil
		} else {
			result[ri.Alias] = u.Object
		}
	}

	return result, nil
}
