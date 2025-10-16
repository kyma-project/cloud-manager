package e2e

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	Get(ctx context.Context, alias string) (client.Object, error)

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

		rObj, err := c.Cluster.GetScheme().New(gvk)
		if err != nil {
			return fmt.Errorf("failed to create object for alias %q GVK %s: %w", decl.Alias, gvk.String(), err)
		}

		src, sourceExists := c.sources[gvk]
		if sourceExists {
			ri.Source = src
			continue
		}

		src = source.Kind(c.Cluster.GetCache(), rObj.(client.Object), handler.Funcs{})
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

func (c *defaultCluster) new(alias string) (client.Object, error) {
	ri, exists := c.resources[alias]
	if !exists {
		return nil, fmt.Errorf("alias %s not found", alias)
	}

	rObj, err := c.Cluster.GetScheme().New(ri.GVK)
	if err != nil {
		return nil, fmt.Errorf("failed to create object for alias %q GVK %s: %w", alias, ri.GVK.String(), err)
	}

	return rObj.(client.Object), nil
}

func (c *defaultCluster) Has(alias string) bool {
	_, ok := c.resources[alias]
	return ok
}

func (c *defaultCluster) Get(ctx context.Context, alias string) (client.Object, error) {
	ri, ok := c.resources[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not found", alias)
	}
	obj, err := c.new(alias)
	if err != nil {
		return nil, err
	}
	err = c.GetClient().Get(ctx, types.NamespacedName{
		Namespace: ri.Namespace,
		Name:      ri.Name,
	}, obj)
	if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error loading resource %q", alias)
	}
	return obj, nil
}

func (c *defaultCluster) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(c.resources))
	for _, ri := range c.resources {
		obj, err := c.new(ri.Alias)
		if err != nil {
			return nil, fmt.Errorf("failed to create object for alias %q: %w", ri.Alias, err)
		}

		err = c.Cluster.GetClient().Get(ctx, types.NamespacedName{
			Namespace: ri.Namespace,
			Name:      ri.Name,
		}, obj)
		if err != nil {
			return nil, fmt.Errorf("failed to load object for alias %q: %w", ri.Alias, err)
		}

		result[ri.Alias] = obj
	}

	return result, nil
}
