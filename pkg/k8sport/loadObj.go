package k8sport

import (
	"context"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sLoadPort interface {
	LoadStateObj(ctx context.Context) error
	LoadObj(ctx context.Context, name types.NamespacedName, obj client.Object) error
	List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error
}

func NewK8sLoadPort(clusterID string) K8sLoadPort {
	return &k8sLoadPortImpl{clusterID: clusterID}
}

func NewK8sLoadPortOnDefaultCluster() K8sLoadPort {
	return NewK8sLoadPort(composedv2.DefaultClusterID)
}

type k8sLoadPortImpl struct {
	clusterID string
}

func (p *k8sLoadPortImpl) LoadStateObj(ctx context.Context) error {
	state := composedv2.StateFromCtx[composedv2.State](ctx)
	return p.LoadObj(ctx, state.Name(), state.Obj())
}

func (p *k8sLoadPortImpl) LoadObj(ctx context.Context, name types.NamespacedName, obj client.Object) error {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	return cluster.K8sClient().Get(ctx, name, obj)
}

func (p *k8sLoadPortImpl) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	return cluster.K8sClient().List(ctx, obj, opts...)
}
