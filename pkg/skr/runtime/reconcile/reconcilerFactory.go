package reconcile

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilerArguments struct {
	KymaRef    klog.ObjectRef
	KcpCluster cluster.Cluster
	SkrCluster cluster.Cluster
	// Provider indicates specific provider resources are available for,
	// if nil all resources for all providers are available
	Provider *cloudcontrolv1beta1.ProviderType

	IgnoreWatchErrors func(bool)
}

type ReconcilerFactory interface {
	New(args ReconcilerArguments) reconcile.Reconciler
}
