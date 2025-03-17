package k8sport

import (
	"context"
	"encoding/json"
	"fmt"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sAnnotateObjPort interface {
	PatchMergeAnnotations(ctx context.Context, obj client.Object, annotations map[string]string) (bool, error)
}

func NewK8sAnnotateObjPort(clusterID string) K8sAnnotateObjPort {
	return &k8sAnnotateObjPortImpl{clusterID: clusterID}
}

func NewK8sAnnotateObjPortOnDefaultCluster() K8sAnnotateObjPort {
	return NewK8sAnnotateObjPort(composedv2.DefaultClusterID)
}

type k8sAnnotateObjPortImpl struct {
	clusterID string
}

func (p *k8sAnnotateObjPortImpl) PatchMergeAnnotations(ctx context.Context, obj client.Object, annotations map[string]string) (bool, error) {
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(make(map[string]string))
	}
	payload := map[string]string{}
	for k, v := range annotations {
		if obj.GetAnnotations()[k] != v {
			payload[k] = v
		}
	}
	if len(payload) == 0 {
		return false, nil
	}
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": payload,
		},
	}
	b, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal annotations for merge patch: %w", err)
	}
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	err = cluster.K8sClient().Patch(ctx, obj, client.RawPatch(types.MergePatchType, b))
	if err == nil {
		for k, v := range annotations {
			obj.GetAnnotations()[k] = v
		}
	}
	return true, err
}
