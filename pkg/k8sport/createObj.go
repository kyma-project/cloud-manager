package k8sport

import (
	"context"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sCreatePort interface {
	Create(ctx context.Context, obj client.Object) error
}

func NewK8sCreatePort(clusterID string) K8sCreatePort {
	return &k8sCreateObjPortImpl{clusterID: clusterID}
}

func NewK8sCreatePortOnDefaultCluster() K8sCreatePort {
	return NewK8sCreatePort(composedv2.DefaultClusterID)
}

var _ K8sCreatePort = &k8sCreateObjPortImpl{}

type k8sCreateObjPortImpl struct {
	clusterID string
}

func (p *k8sCreateObjPortImpl) Create(ctx context.Context, obj client.Object) error {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	return cluster.K8sClient().Create(ctx, obj)
}
