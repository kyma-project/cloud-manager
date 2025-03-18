package k8sport

import (
	"context"
	"encoding/json"
	"fmt"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sLabelObjPort interface {
	PatchMergeLabels(ctx context.Context, obj client.Object, labels map[string]string) (bool, error)
}

func NewK8sLabelObjPort(clusterID string) K8sLabelObjPort {
	return &k8sLabelObjPortImpl{clusterID: clusterID}
}

func NewK8sLabelObjPortOnDefaultCluster() K8sLabelObjPort {
	return NewK8sLabelObjPort(composedv2.DefaultClusterID)
}

type k8sLabelObjPortImpl struct {
	clusterID string
}

func (p *k8sLabelObjPortImpl) PatchMergeLabels(ctx context.Context, obj client.Object, labels map[string]string) (bool, error) {
	if obj.GetLabels() == nil {
		obj.SetLabels(make(map[string]string))
	}
	payload := map[string]string{}
	for k, v := range labels {
		if obj.GetLabels()[k] != v {
			payload[k] = v
		}
	}
	if len(payload) == 0 {
		return false, nil
	}
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": payload,
		},
	}
	b, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal labels for merge patch: %w", err)
	}
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	err = cluster.K8sClient().Patch(ctx, obj, client.RawPatch(types.MergePatchType, b))
	if err == nil {
		for k, v := range labels {
			obj.GetLabels()[k] = v
		}
	}
	return true, err
}
