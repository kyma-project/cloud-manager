package sim

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

// Start starts Runtime, GardenerCluster and KymaKcp managers and blocks until the context is done.
func Start(ctx context.Context, kcp cluster.Cluster, garden cluster.Cluster, skrProvider SkrProvider, logger logr.Logger) error {
	mngr := NewManager(kcp, logger)
	var err error

	err = NewSimRuntime(mngr, kcp.GetClient(), garden.GetClient())
	if err != nil {
		return fmt.Errorf("could not create runtime manager: %w", err)
	}

	err = NewSimGardenerCluster(mngr, kcp.GetClient(), garden.GetClient())
	if err != nil {
		return fmt.Errorf("could not create gardener cluster manager: %w", err)
	}

	err = NewSimKymaKcp(mngr, kcp.GetClient(), skrProvider)
	if err != nil {
		return fmt.Errorf("could not create Kyma KCP manager: %w", err)
	}

	err = mngr.Start(ctx)
	if err != nil {
		return fmt.Errorf("error running sim manager: %w", err)
	}
	return nil
}
