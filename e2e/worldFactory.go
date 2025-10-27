package e2e

import (
	"context"
	"fmt"
	"sync"
	"time"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type WorldFactory struct {
}

func NewWorldFactory() *WorldFactory {
	return &WorldFactory{}
}

type WorldCreateOptions struct {
	KcpRestConfig         *rest.Config
	CloudProfileLoader    sim.CloudProfileLoader
	SkrKubeconfigProvider sim.SkrKubeconfigProvider
	ExtraRunnables        []manager.Runnable
}

func (f *WorldFactory) Create(rootCtx context.Context, opts WorldCreateOptions) (WorldIntf, error) {
	factoryKcp := NewKcpClusterFactory(opts.KcpRestConfig)
	kcpCluster, err := factoryKcp.CreateCluster(rootCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kcp cluster: %w", err)
	}

	waitClusterStarts := func(ctx context.Context, c cluster.Cluster) bool {
		toCtx, toCancel := context.WithTimeout(ctx, 10*time.Second)
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
		if err := kcpCluster.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running KCP cluster: %w", err))
		}
	}()

	if !waitClusterStarts(ctx, kcpCluster) {
		cancel()
		return nil, fmt.Errorf("kcp cache did not sync")
	}

	if err := InitializeKcp(ctx, kcpCluster.GetClient()); err != nil {
		cancel()
		return nil, fmt.Errorf("error initializing KCP cluster: %w", err)
	}

	time.Sleep(time.Second)

	// Garden Cluster ------------------------

	factoryGarden := NewGardenClusterFactory(kcpCluster.GetClient())
	gardenCluster, err := factoryGarden.CreateCluster(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creating garden cluster: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := gardenCluster.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running Garden cluster: %w", err))
		}
	}()

	if !waitClusterStarts(ctx, gardenCluster) {
		cancel()
		return nil, fmt.Errorf("garden cache did not sync")
	}

	err = gardenCluster.AddResources(ctx, &ResourceDeclaration{
		Alias:      "dummy-foo-just-starting-shoot-watch",
		Kind:       "Shoot",
		ApiVersion: gardenerapicore.SchemeGroupVersion.String(),
		Name:       "none",
		Namespace:  e2econfig.Config.GardenNamespace,
	})
	if err != nil {
		return result, fmt.Errorf("error adding dummy Shoot resource declaration: %w", err)
	}

	// SIM -----------------------------------

	time.Sleep(time.Second)

	if opts.CloudProfileLoader == nil {
		opts.CloudProfileLoader = sim.NewGardenCloudProfileLoader(gardenCluster.GetClient(), e2econfig.Config.GardenNamespace)
	}
	if opts.SkrKubeconfigProvider == nil {
		opts.SkrKubeconfigProvider = sim.NewGardenSkrKubeconfigProvider(gardenCluster.GetClient(), 10*time.Hour)
	}

	simInstance, err := sim.New(sim.CreateOptions{
		KCP:                   kcpCluster,
		Garden:                gardenCluster.GetClient(),
		CloudProfileLoader:    opts.CloudProfileLoader,
		Logger:                ctrl.Log,
		SkrKubeconfigProvider: opts.SkrKubeconfigProvider,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creating sim instance: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := simInstance.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running sim instance: %w", err))
		}
	}()

	for _, r := range opts.ExtraRunnables {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Start(ctx); err != nil {
				result.runError = multierror.Append(result.runError, fmt.Errorf("error running extra runnable %T: %w", r, err))
			}
		}()
	}

	result.kcp = kcpCluster
	result.garden = gardenCluster
	result.simu = simInstance

	time.Sleep(time.Second)

	return result, nil
}
