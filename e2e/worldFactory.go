package e2e

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/e2e/cloud"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	"k8s.io/client-go/rest"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type WorldFactory struct {
}

func NewWorldFactory() *WorldFactory {
	return &WorldFactory{}
}

type WorldCreateOptions struct {
	Config                *e2econfig.ConfigType
	KcpRestConfig         *rest.Config
	CloudProfileLoader    e2elib.CloudProfileLoader
	SkrKubeconfigProvider e2elib.SkrKubeconfigProvider
	SkrManagerFactory     e2ekeb.SkrManagerFactory
}

func (f *WorldFactory) Create(rootCtx context.Context, opts WorldCreateOptions) (WorldIntf, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required: %w", common.ErrLogical)
	}
	factoryKcp := e2elib.NewKcpClientFactory(opts.KcpRestConfig)
	kcpManager, err := factoryKcp.CreateManager(rootCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kcp cluster manager: %w", err)
	}

	waitClusterStarts := func(ctx context.Context, c cluster.Cluster) bool {
		var toCtx context.Context
		var toCancel context.CancelFunc
		if debugged.Debugged {
			toCtx, toCancel = context.WithTimeout(ctx, 10*time.Minute)
		} else {
			toCtx, toCancel = context.WithTimeout(ctx, 10*time.Second)
		}
		defer toCancel()
		if ok := c.GetCache().WaitForCacheSync(toCtx); !ok {
			return false
		}
		return true
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(rootCtx)
	result := &defaultWorld{
		mCtx:   ctx,
		cancel: cancel,
		wg:     &wg,
	}

	// KCP Cluster ----------------------------

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kcpManager.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running KCP cluster manager: %w", err))
		}
	}()

	if !waitClusterStarts(ctx, kcpManager) {
		cancel()
		return nil, fmt.Errorf("kcp cache did not sync")
	}

	if err := e2elib.InitializeKcp(ctx, kcpManager.GetClient(), opts.Config); err != nil {
		cancel()
		return nil, fmt.Errorf("error initializing KCP cluster: %w", err)
	}

	time.Sleep(time.Second)

	// Garden Cluster ------------------------

	factoryGarden := e2elib.NewGardenClientFactory(kcpManager.GetClient(), opts.Config.GardenNamespace)
	gardenManager, err := factoryGarden.CreateManager(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creating garden cluster manager: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := gardenManager.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running Garden cluster: %w", err))
		}
	}()

	if !waitClusterStarts(ctx, gardenManager) {
		cancel()
		return nil, fmt.Errorf("garden cache did not sync")
	}

	time.Sleep(time.Second)

	if opts.CloudProfileLoader == nil {
		opts.CloudProfileLoader = e2elib.NewGardenCloudProfileLoader(gardenManager.GetClient(), opts.Config)
	}
	if opts.SkrKubeconfigProvider == nil {
		opts.SkrKubeconfigProvider = e2elib.NewGardenSkrKubeconfigProvider(gardenManager.GetClient(), 10*time.Hour, opts.Config.GardenNamespace)
	}
	if opts.SkrManagerFactory == nil {
		opts.SkrManagerFactory = e2ekeb.NewSkrManagerFactory(kcpManager.GetClient(), clock.RealClock{}, opts.Config.KcpNamespace)
	}

	// KEB -----------------------------------

	kebi := e2ekeb.NewKeb(kcpManager.GetClient(), gardenManager.GetClient(), opts.SkrManagerFactory, opts.CloudProfileLoader, opts.SkrKubeconfigProvider, opts.Config)

	// give time to sim reconciler kinds to get added to cache and start syncing
	time.Sleep(time.Second)

	if !waitClusterStarts(ctx, kcpManager) {
		cancel()
		return nil, fmt.Errorf("kcp cache did not sync after sim creation")
	}

	// Cloud -------------------------------

	cl, err := cloud.Create(ctx, gardenManager.GetClient(), opts.Config)
	if err != nil {
		return nil, fmt.Errorf("error creating cloud: %w", err)
	}

	result.config = opts.Config
	result.kcpManager = kcpManager
	result.gardenManager = gardenManager
	result.kebi = kebi
	result.cloud = cl
	result.kcp = NewCluster(ctx, "kcp", kcpManager)
	result.garden = NewCluster(ctx, "garden", gardenManager)

	return result, nil
}
