package main

import (
	"os"

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

	_ = e2econfig.LoadConfig()

	cp := e2e.NewClusterProvider()
	defer func() {
		if err := cp.Stop(); err != nil {
			logger.Error(err, "error stopping clusterProvider")
		}
	}()

	if err := cp.Init(ctx); err != nil {
		logger.Error(err, "error initializing clusterProvider")
		os.Exit(1)
	}

	kcp, err := cp.KCP(ctx)
	if err != nil {
		logger.Error(err, "Failed to get KCP provider")
		os.Exit(1)
		return
	}
	garden, err := cp.Garden(ctx)
	if err != nil {
		logger.Error(err, "Failed to get Gardener provider")
		os.Exit(1)
		return
	}

	simu, err := sim.New(kcp, garden, logger)
	if err != nil {
		logger.Error(err, "Failed to create sim")
		os.Exit(1)
		return
	}
	if err := simu.Start(ctx); err != nil {
		logger.Error(err, "Failed running sim")
		os.Exit(1)
	}
}
