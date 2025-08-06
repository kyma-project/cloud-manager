package e2e

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func NewCluster(clstr cluster.Cluster) *Cluster {
	return &Cluster{
		Cluster:   clstr,
		resources: make(map[string]*ResourceInfo),
		sources:   make(map[schema.GroupVersionKind]source.SyncingSource),
	}
}

type Cluster struct {
	cluster.Cluster

	runCtx     context.Context
	cancelRun  context.CancelFunc
	started    bool
	stoppingCh chan error

	resources map[string]*ResourceInfo
	sources   map[schema.GroupVersionKind]source.SyncingSource
}

func (c *Cluster) IsStarted() bool {
	return c.started
}

func (c *Cluster) Start(ctx context.Context) error {
	if c.started {
		return fmt.Errorf("cluster already started")
	}

	c.runCtx, c.cancelRun = context.WithCancel(ctx)
	c.stoppingCh = make(chan error)

	var err error
	go func() {
		err = c.Cluster.Start(c.runCtx)
		c.stoppingCh <- err
		close(c.stoppingCh)
	}()

	// Wait for the cluster to be started... and as a side effect informers synced
	ok := c.Cluster.GetCache().WaitForCacheSync(c.runCtx)

	if err != nil {
		err = multierror.Append(err, fmt.Errorf("failed to start cluster: %w", err))
	}
	if !ok {
		err = multierror.Append(err, fmt.Errorf("failed to wait for cache sync"))
	}

	c.started = true

	return err
}

func (c *Cluster) Stop() error {
	if !c.started {
		return fmt.Errorf("cluster not started")
	}
	c.cancelRun()
	err := <-c.stoppingCh
	c.started = false
	return err
}

func (c *Cluster) AddResources(ctx context.Context, arr []*ResourceDeclaration) error {
	if !c.started {
		return fmt.Errorf("cluster not started")
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

func (c *Cluster) New(alias string) (client.Object, error) {
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

func (c *Cluster) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(c.resources))
	for _, ri := range c.resources {
		obj, err := c.New(ri.Alias)
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
