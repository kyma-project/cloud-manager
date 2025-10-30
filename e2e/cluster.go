package e2e

import (
	"context"
	"fmt"
	"sync"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ClusterEvaluationHandle interface {
	ClusterAlias() string

	// AllResources returns all declared resources in undefined order that might not be the same to the order they were declared
	AllResources() []*ResourceInfo

	// GetResource returns resource declaration for the given alias
	GetResource(alias string) *ResourceInfo

	// Get returns resource declared with the given alias from the cache. If resource does not exist, the nil object
	// and no error are returned. Also, if kind is not registered, nil resource and no error are returned.
	// Error is returned only if it's some other than NotFound and NoMatch
	Get(ctx context.Context, alias string) (map[string]interface{}, error)

	RestMapping(alias string) (*meta.RESTMapping, error)
}

type Cluster interface {
	cluster.Cluster
	ClusterEvaluationHandle

	// AddResources declares k8s objects that are watched and can be got from the cache
	AddResources(ctx context.Context, arr ...*ResourceDeclaration) error

	// EvaluationContext returns map of all declared resources
	//EvaluationContext(ctx context.Context) (map[string]interface{}, error)
}

func NewCluster(alias string, clstr cluster.Cluster) Cluster {
	return &defaultCluster{
		Cluster:      clstr,
		clusterAlias: alias,
		resources:    make(map[string]*ResourceInfo),
		sources:      make(map[schema.GroupVersionKind]source.SyncingSource),
		mappingCache: make(map[string]*meta.RESTMapping),
	}
}

type defaultCluster struct {
	m sync.Mutex

	cluster.Cluster

	clusterAlias string

	runCtx context.Context

	resources map[string]*ResourceInfo
	sources   map[schema.GroupVersionKind]source.SyncingSource

	mappingCache map[string]*meta.RESTMapping
}

func (c *defaultCluster) Start(ctx context.Context) error {
	c.runCtx = ctx
	return c.Cluster.Start(ctx)
}

func (c *defaultCluster) ClusterAlias() string {
	return c.clusterAlias
}

func (c *defaultCluster) AllResources() []*ResourceInfo {
	return pie.Values(c.resources)
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

func (c *defaultCluster) GetResource(alias string) *ResourceInfo {
	return c.resources[alias]
}

func (c *defaultCluster) RestMapping(alias string) (*meta.RESTMapping, error) {
	ri, ok := c.resources[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not declared", alias)
	}
	c.m.Lock()
	defer c.m.Unlock()
	if m, ok := c.mappingCache[ri.GVK.String()]; ok {
		return m, nil
	}

	mapping, err := c.GetRESTMapper().RESTMapping(ri.GVK.GroupKind(), ri.GVK.Version)
	if util.IgnoreNoMatch(err) != nil {
		return nil, fmt.Errorf("error getting rest mapping for %q: %w", alias, err)
	}
	if err != nil {
		return nil, nil
	}
	c.mappingCache[ri.GVK.String()] = mapping
	return mapping, nil
}

func (c *defaultCluster) Get(ctx context.Context, alias string) (map[string]interface{}, error) {
	ri, ok := c.resources[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not declared", alias)
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
	return u.Object, nil
}

//func (c *defaultCluster) EvaluationContext(ctx context.Context) (map[string]interface{}, error) {
//	data, err := c.evaluationContextWithAlreadyEvaluated(ctx)
//	if err != nil {
//		return data, err
//	}
//	eval := NewEvaluator()
//	if err := eval.SetGlobalVars(data); err != nil {
//		return nil, fmt.Errorf("failed to set global vars: %w", err)
//	}
//
//	var newlyEvaluated []string
//	for _, ri := range c.resources {
//		if ri.Evaluated {
//			continue
//		}
//		err = eval.EvalResource(ri)
//		if err == nil && ri.Evaluated {
//			newlyEvaluated = append(newlyEvaluated, ri.Name)
//		}
//	}
//
//	return data, err
//}
//
//func (c *defaultCluster) recursiveEval(ctx context.Context, eval Evaluator, result map[string]interface{}) (map[string]interface{}, error) {
//	if err := eval.SetGlobalVars(result); err != nil {
//		return nil, fmt.Errorf("failed to set global vars: %w", err)
//	}
//	// load already evaluated resources
//	for _, ri := range c.resources {
//		if !ri.Evaluated {
//			// try to evaluate it
//			err := eval.EvalResource(ri)
//			if err == nil && ri.Evaluated {
//				// successfully evaluated
//				// try to load it
//				data, err := c.Get(ctx, ri.Name)
//				if err != nil {
//					return nil, err
//				}
//				result[ri.Name] = data
//				if err := eval.SetGlobalVar(ri.Name, data); err != nil {
//					return nil, fmt.Errorf("failed to set global var %q: %w", ri.Name, err)
//				}
//			}
//		}
//	}
//}
//
//func (c *defaultCluster) evaluationContextWithAlreadyEvaluated(ctx context.Context) (map[string]interface{}, error) {
//	result := make(map[string]interface{}, len(c.resources))
//	for _, ri := range c.resources {
//		var data map[string]interface{}
//		if ri.Evaluated {
//			u := &unstructured.Unstructured{}
//			u.SetGroupVersionKind(ri.GVK)
//
//			err := c.Cluster.GetClient().Get(ctx, types.NamespacedName{
//				Namespace: ri.Namespace,
//				Name:      ri.Name,
//			}, u)
//			if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
//				return nil, fmt.Errorf("failed to load object for alias %q: %w", ri.Alias, err)
//			}
//			if err == nil {
//				data = u.Object
//			}
//		}
//		result[ri.Alias] = data
//	}
//
//	return result, nil
//}
