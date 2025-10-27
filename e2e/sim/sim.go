package sim

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	DoNotReconcile = "sim.do-not-reconcile"
)

type Sim interface {
	ClientClusterFactory
	// Start starts Runtime, GardenerCluster and KymaKcp managers and blocks until the context is done.
	Start(ctx context.Context) error
	Keb() Keb
}

type CreateOptions struct {
	KCP                   cluster.Cluster
	Garden                client.Client
	CloudProfileLoader    CloudProfileLoader
	Logger                logr.Logger
	SkrKubeconfigProvider SkrKubeconfigProvider
}

func (o *CreateOptions) Validate() error {
	var result error
	if o.KCP == nil {
		result = multierror.Append(fmt.Errorf("missing KCP cluster"))
	}
	if o.Garden == nil {
		result = multierror.Append(fmt.Errorf("missing Garden cluster"))
	}
	if o.Logger.GetSink() == nil {
		result = multierror.Append(fmt.Errorf("missing Logger cluster"))
	}
	if o.CloudProfileLoader == nil {
		result = multierror.Append(fmt.Errorf("missing CloudProfileLoader"))
	}
	if o.SkrKubeconfigProvider == nil {
		result = multierror.Append(fmt.Errorf("missing SkrKubeconfigProvider"))
	}
	return result
}

func New(opts CreateOptions) (Sim, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid new sim options: %w", err)
	}

	if opts.SkrKubeconfigProvider == nil {
		opts.SkrKubeconfigProvider = NewGardenSkrKubeconfigProvider(opts.Garden, 10*time.Hour)
	}

	mngr := NewManager(opts.KCP, opts.Logger)

	factory := NewClientClusterFactory(opts.KCP.GetClient(), clock.RealClock{})

	keb := NewKeb(opts.KCP.GetClient(), opts.Garden, opts.CloudProfileLoader)

	simRt := newSimRuntime(opts.KCP.GetClient(), opts.Garden, opts.CloudProfileLoader)
	if err := simRt.SetupWithManager(mngr); err != nil {
		return nil, fmt.Errorf("could not create runtime manager: %w", err)
	}

	simGC := newSimGardenerCluster(opts.KCP.GetClient(), opts.SkrKubeconfigProvider)
	if err := simGC.SetupWithManager(mngr); err != nil {
		return nil, fmt.Errorf("could not create gardener cluster manager: %w", err)
	}

	simKK := newSimKymaKcp(opts.KCP.GetClient(), factory)
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
	s.simKK.mainCtx = ctx
	if err := s.mngr.Start(ctx); err != nil {
		return fmt.Errorf("error running sim manager: %w", err)
	}
	return nil
}
