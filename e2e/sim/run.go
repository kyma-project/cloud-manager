package sim

import (
	"context"
	"fmt"
	"time"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Run(ctx context.Context, config *e2econfig.ConfigType) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	kcpClientFactory := e2elib.NewKcpClientFactory(ctrl.GetConfigOrDie())
	kcpManager, err := kcpClientFactory.CreateManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kcp manager: %w", err)
	}

	runCh := make(chan error)
	go func() {
		err := kcpManager.Start(ctx)
		runCh <- err
	}()

	if err := func() error {
		toCtx, toCancel := context.WithTimeout(ctx, 10*time.Second)
		defer toCancel()
		if !kcpManager.GetCache().WaitForCacheSync(toCtx) {
			return fmt.Errorf("failed to sync kcp manager cache")
		}
		return nil
	}(); err != nil {
		return err
	}

	// kcp manager is started and synced

	if err := e2elib.InitializeKcp(ctx, kcpManager.GetClient(), config); err != nil {
		return fmt.Errorf("failed to initialize kcp: %w", err)
	}

	// kcp cluster is initialized

	if !kcpManager.GetCache().WaitForCacheSync(ctx) {
		return fmt.Errorf("failed to sync kcp manager cache after initialization")
	}

	gardenClientFactory := e2elib.NewGardenClientFactory(kcpManager.GetClient(), config.GardenNamespace)
	gardenClient, err := gardenClientFactory.CreateClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create garden client: %w", err)
	}

	cpl := e2elib.NewGardenCloudProfileLoader(gardenClient, config)
	skrKubeconfigProvider := e2elib.NewGardenSkrKubeconfigProvider(gardenClient, 10*time.Hour, config.GardenNamespace)
	skrManagerFactory := e2ekeb.NewSkrManagerFactory(kcpManager.GetClient(), clock.RealClock{}, config.KcpNamespace)

	opts := CreateOptions{
		Config:                config,
		StartCtx:              ctx,
		KcpManager:            kcpManager,
		GardenClient:          gardenClient,
		Logger:                ctrl.Log,
		CloudProfileLoader:    cpl,
		SkrKubeconfigProvider: skrKubeconfigProvider,
		SkrManagerFactory:     skrManagerFactory,
	}
	_, err = New(opts)
	if err != nil {
		return fmt.Errorf("failed creating sim: %w", err)
	}

	// block until kcpManager stops and writes to chan

	err = <-runCh
	return err
}
