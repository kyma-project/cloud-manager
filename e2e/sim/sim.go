package sim

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	DoNotReconcile = "sim.do-not-reconcile"
)

type Sim interface {
	ClientClusterFactory
	Start(ctx context.Context) error
	Keb() Keb
}

func New(kcp cluster.Cluster, garden cluster.Cluster, logger logr.Logger) (Sim, error) {
	mngr := NewManager(kcp, logger)

	factory := NewClientClusterFactory(kcp.GetClient())

	keb := NewKeb(kcp.GetClient())

	simRt := newSimRuntime(kcp.GetClient(), garden.GetClient())
	if err := simRt.SetupWithManager(mngr); err != nil {
		return nil, fmt.Errorf("could not create runtime manager: %w", err)
	}

	simGC := newSimGardenerCluster(kcp.GetClient(), garden.GetClient())
	if err := simGC.SetupWithManager(mngr); err != nil {
		return nil, fmt.Errorf("could not create gardener cluster manager: %w", err)
	}

	simKK := newSimKymaKcp(kcp.GetClient(), factory)
	if err := simKK.SetupWithManager(mngr); err != nil {
		return nil, fmt.Errorf("could not create Kyma KCP manager: %w", err)
	}

	return &defaultSim{
		ClientClusterFactory: factory,
		mngr:                 mngr,
		keb:                  keb,
		simRT:                simRt,
		simGC:                simGC,
		simKK:                simKK,
	}, nil
}

type defaultSim struct {
	ClientClusterFactory

	mngr  manager.Manager
	keb   Keb
	simRT *simRuntime
	simGC *simGardenerCluster
	simKK *simKymaKcp
}

func (s *defaultSim) Keb() Keb {
	return s.keb
}

// Start starts Runtime, GardenerCluster and KymaKcp managers and blocks until the context is done.
func (s *defaultSim) Start(ctx context.Context) error {
	if err := s.mngr.Start(ctx); err != nil {
		return fmt.Errorf("error running sim manager: %w", err)
	}
	return nil
}
