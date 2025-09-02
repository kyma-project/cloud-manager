package sim

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type SkrProvider interface {
	GetSKR(ctx context.Context, runtimeID string) (cluster.Cluster, error)
}
