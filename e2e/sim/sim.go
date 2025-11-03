package sim

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	DoNotReconcile = "sim.do-not-reconcile"
)

type Sim interface {
	SkrManagerFactory
	Keb() Keb
}

type CreateOptions struct {
	Config                *e2econfig.ConfigType
	StartCtx              context.Context
	KcpManager            manager.Manager
	Garden                client.Client
	GardenApiReader       client.Reader
	CloudProfileLoader    CloudProfileLoader
	Logger                logr.Logger
	SkrKubeconfigProvider SkrKubeconfigProvider
}

func (o *CreateOptions) Validate() error {
	var result error
	if o.Config == nil {
		result = multierror.Append(fmt.Errorf("missing Config"))
	}
	if o.StartCtx == nil {
		result = multierror.Append(fmt.Errorf("missing StartCtx"))
	}
	if o.KcpManager == nil {
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
		opts.SkrKubeconfigProvider = NewGardenSkrKubeconfigProvider(opts.Garden, 10*time.Hour, opts.Config.GardenNamespace)
	}

	factory := NewClientClusterFactory(opts.KcpManager.GetClient(), clock.RealClock{}, opts.Config.KcpNamespace)

	keb := NewKeb(opts.KcpManager.GetClient(), opts.Garden, opts.CloudProfileLoader, opts.Config)

	simRt := newSimRuntime(opts.KcpManager.GetClient(), opts.Garden, opts.CloudProfileLoader, opts.GardenApiReader, opts.Config)
	if err := simRt.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create runtime manager: %w", err)
	}

	simGC := newSimGardenerCluster(opts.KcpManager.GetClient(), opts.SkrKubeconfigProvider)
	if err := simGC.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create gardener cluster manager: %w", err)
	}

	simKK := newSimKymaKcp(opts.StartCtx, opts.KcpManager.GetClient(), factory, opts.Config.KcpNamespace)
	if err := simKK.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create Kyma KCP manager: %w", err)
	}

	return &defaultSim{
		SkrManagerFactory: factory,
		keb:               keb,
		simRT:             simRt,
		simGC:             simGC,
		simKK:             simKK,
	}, nil
}

type defaultSim struct {
	SkrManagerFactory

	keb   Keb
	simRT *simRuntime
	simGC *simGardenerCluster
	simKK *simKymaKcp
}

func (s *defaultSim) Keb() Keb {
	return s.keb
}
