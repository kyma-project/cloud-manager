package e2e

import (
	"context"

	"github.com/elliotchance/pie/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type clusterEvaluationHandleFake struct {
	clusterAlias string
	resources    map[string]*ResourceInfo
	items        map[string]map[string]interface{}
}

func newClusterEvaluationHandleFake(clusterAlias string) *clusterEvaluationHandleFake {
	return &clusterEvaluationHandleFake{
		clusterAlias: clusterAlias,
		resources:    make(map[string]*ResourceInfo),
		items:        make(map[string]map[string]interface{}),
	}
}

func (h *clusterEvaluationHandleFake) setObj(alias string, obj interface{}) {
	if obj == nil {
		h.items[alias] = nil
		return
	}
	mm, ok := obj.(map[string]interface{})
	if ok {
		h.items[alias] = mm
		return
	}
	mm, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	h.items[alias] = mm
}

func (h *clusterEvaluationHandleFake) declare(alias, name, namespace string) {
	h.resources[alias] = &ResourceInfo{
		ResourceDeclaration: ResourceDeclaration{
			Alias:      alias,
			Kind:       "SomeKind",
			ApiVersion: "v1",
			Name:       name,
			Namespace:  namespace,
		},
	}
}

func (h *clusterEvaluationHandleFake) ClusterAlias() string {
	return h.clusterAlias
}

func (h *clusterEvaluationHandleFake) AllResources() []*ResourceInfo {
	return pie.Values(h.resources)
}

func (h *clusterEvaluationHandleFake) GetResource(alias string) *ResourceInfo {
	return h.resources[alias]
}

func (h *clusterEvaluationHandleFake) Get(_ context.Context, alias string) (map[string]interface{}, error) {
	return h.items[alias], nil
}

func (h *clusterEvaluationHandleFake) RestMapping(alias string) (*meta.RESTMapping, error) {
	return nil, nil
}
