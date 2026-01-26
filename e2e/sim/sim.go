package sim

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Sim interface {
	Sim() bool
}

type CreateOptions struct {
	// required

	Config *e2econfig.ConfigType

	// StartCtx must be the same context that's given to the KcpManager.Start() so sim can stop
	// its controllers once the KcpManager is stopped
	StartCtx              context.Context
	KcpManager            manager.Manager
	GardenClient          client.Client
	Logger                logr.Logger
	CloudProfileLoader    e2elib.CloudProfileLoader
	SkrKubeconfigProvider e2elib.SkrKubeconfigProvider

	// optional
	SkrManagerFactory e2ekeb.SkrManagerFactory
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
	if o.GardenClient == nil {
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

	if opts.SkrManagerFactory == nil {
		opts.SkrManagerFactory = e2ekeb.NewSkrManagerFactory(opts.KcpManager.GetClient(), clock.RealClock{}, opts.Config.KcpNamespace)
	}

	simRt := newSimRuntime(opts.KcpManager.GetClient(), opts.GardenClient, opts.CloudProfileLoader, opts.Config)
	if err := simRt.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create runtime manager: %w", err)
	}

	simGC := newSimGardenerCluster(opts.KcpManager.GetClient(), opts.SkrKubeconfigProvider)
	if err := simGC.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create gardener cluster manager: %w", err)
	}

	simKK := newSimKymaKcp(opts.StartCtx, opts.KcpManager.GetClient(), opts.SkrManagerFactory, opts.Config.KcpNamespace)
	if err := simKK.SetupWithManager(opts.KcpManager); err != nil {
		return nil, fmt.Errorf("could not create Kyma KCP manager: %w", err)
	}

	return &defaultSim{
		simRT: simRt,
		simGC: simGC,
		simKK: simKK,
	}, nil
}

type defaultSim struct {
	simRT *simRuntime
	simGC *simGardenerCluster
	simKK *simKymaKcp
}

func (s *defaultSim) Sim() bool {
	return true
}
