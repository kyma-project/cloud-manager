package reconcile

import (
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerArguments struct {
	KymaRef    klog.ObjectRef
	KcpCluster cluster.Cluster
	SkrCluster cluster.Cluster
}

type ReconcilerFactory interface {
	New(args ReconcilerArguments) reconcile.Reconciler
}
