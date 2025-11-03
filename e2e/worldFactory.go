package e2e

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
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
	CloudProfileLoader    sim.CloudProfileLoader
	SkrKubeconfigProvider sim.SkrKubeconfigProvider
}

func (f *WorldFactory) Create(rootCtx context.Context, opts WorldCreateOptions) (WorldIntf, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required: %w", common.ErrLogical)
	}
	factoryKcp := NewKcpClusterManagerFactory(opts.KcpRestConfig)
	kcpManager, err := factoryKcp.CreateClusterManager(rootCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kcp cluster manager: %w", err)
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
		if err := kcpManager.Start(ctx); err != nil {
			result.runError = multierror.Append(result.runError, fmt.Errorf("error running KCP cluster manager: %w", err))
		}
	}()

	if !waitClusterStarts(ctx, kcpManager) {
		cancel()
		return nil, fmt.Errorf("kcp cache did not sync")
	}

	cmInfo := &corev1.ConfigMap{}
	err = kcpManager.GetAPIReader().Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      "cm-info",
	}, cmInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap info: %w", err)
	}
	fmt.Println("WorldFactory cm-info for KCP ==================================")
	fmt.Printf("%+v\n", cmInfo.Data)

	if err := InitializeKcp(ctx, kcpManager.GetClient(), opts.Config); err != nil {
		cancel()
		return nil, fmt.Errorf("error initializing KCP cluster: %w", err)
	}

	time.Sleep(time.Second)

	// Garden Cluster ------------------------

	factoryGarden := NewGardenClusterManagerFactory(kcpManager.GetClient(), opts.Config.GardenNamespace)
	gardenManager, err := factoryGarden.CreateClusterManager(ctx)
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

	cmInfo = &corev1.ConfigMap{}
	err = gardenManager.GetAPIReader().Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      "cm-info",
	}, cmInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap info: %w", err)
	}
	fmt.Println("WorldFactory cm-info for Garden ==================================")
	fmt.Printf("%+v\n", cmInfo.Data)

	// SIM -----------------------------------

	time.Sleep(time.Second)

	if opts.CloudProfileLoader == nil {
		opts.CloudProfileLoader = sim.NewGardenCloudProfileLoader(gardenManager.GetClient(), opts.Config)
	}
	if opts.SkrKubeconfigProvider == nil {
		opts.SkrKubeconfigProvider = sim.NewGardenSkrKubeconfigProvider(gardenManager.GetClient(), 10*time.Hour, opts.Config.GardenNamespace)
	}

	simInstance, err := sim.New(sim.CreateOptions{
		Config:                opts.Config,
		StartCtx:              ctx,
		KcpManager:            kcpManager,
		Garden:                gardenManager.GetClient(),
		GardenApiReader:       gardenManager.GetAPIReader(),
		CloudProfileLoader:    opts.CloudProfileLoader,
		Logger:                ctrl.Log,
		SkrKubeconfigProvider: opts.SkrKubeconfigProvider,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creating sim instance: %w", err)
	}

	result.config = opts.Config
	result.kcpManager = kcpManager
	result.gardenManager = gardenManager
	result.simu = simInstance
	result.kcp = NewCluster(ctx, "kcp", kcpManager)
	result.garden = NewCluster(ctx, "garden", gardenManager)

	return result, nil
}
