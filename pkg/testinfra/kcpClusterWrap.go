package testinfra

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

func newKcpClusterWrap(original cluster.Cluster) *kcpClusterWrap {
	return &kcpClusterWrap{
		Cluster: original,
		delegated: &clientWithDelegatedReader{
			Client:          original.GetClient(),
			delegatedReader: original.GetAPIReader(),
		},
	}
}

type kcpClusterWrap struct {
	cluster.Cluster
	delegated *clientWithDelegatedReader
}

func (w *kcpClusterWrap) GetClient() client.Client {
	return w.delegated
}
