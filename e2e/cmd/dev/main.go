package main

import (
	"fmt"
	"os"
	"path"

	"github.com/kyma-project/cloud-manager/e2e"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	ctx := ctrl.SetupSignalHandler()

	opts := zap.Options{}
	opts.Development = true
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	cfg := e2econfig.LoadConfig()
	sharedStateFile := path.Join(cfg.GetBaseDir(), ".runtimes.yaml")
	sharedState, err := e2e.LoadSharedState(sharedStateFile)
	if err != nil {
		panic(fmt.Errorf("failed loading shared runtimes state: %w", err))
	}

	f := e2e.NewWorldFactory()
	world, err := f.Create(ctx)
	if err != nil {
		panic(err)
	}

	for _, runtimeId := range sharedState.Runtimes {
		logger.WithValues("runtimeID", runtimeId).Info("Importing runtime...")
		skr, err := world.SKR().ImportShared(ctx, runtimeId)
		if err != nil {
			panic(fmt.Errorf("failed importing shared runtime %s: %w", runtimeId, err))
		}

		logger.WithValues(
			"runtimeID", runtimeId,
			"shoot", skr.ShootName,
			"provider", skr.Provider,
			"alias", skr.Alias,
		).Info("Shared runtime imported")
	}

	kcp, err := world.ClusterProvider().KCP(ctx)
	if err != nil {
		logger.Error(err, "Failed to get KCP provider")
		os.Exit(1)
		return
	}
	garden, err := world.ClusterProvider().Garden(ctx)
	if err != nil {
		logger.Error(err, "Failed to get Gardener provider")
		os.Exit(1)
		return
	}
	skrProvider := e2e.NewSkrProvider(kcp.GetClient(), world.SKR())
	if err := sim.Start(ctx, kcp, garden, skrProvider, logger); err != nil {
		logger.Error(err, "Failed running sim")
		os.Exit(1)
	}
}
