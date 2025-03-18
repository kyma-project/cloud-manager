package k8sport

import (
	"context"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sDeletePort interface {
	Delete(ctx context.Context, obj client.Object) error
}

func NewK8sDeletePort(clusterID string) K8sDeletePort {
	return &k8sDeleteObjectPortImpl{clusterID: clusterID}
}

var _ K8sDeletePort = &k8sDeleteObjectPortImpl{}

type k8sDeleteObjectPortImpl struct {
	clusterID string
}

func (p *k8sDeleteObjectPortImpl) Delete(ctx context.Context, obj client.Object) error {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	return cluster.K8sClient().Delete(ctx, obj)
}
