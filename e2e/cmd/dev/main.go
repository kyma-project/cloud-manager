package main

import (
	"context"
	"fmt"
	"os"
	"time"

	//cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/e2e"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer cancel()

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

	simu, err := sim.New(ctx, sim.CreateOptions{
		KCP:    kcp,
		Garden: garden,
		Logger: logger,
	})
	if err != nil {
		logger.Error(err, "Failed to create sim")
		os.Exit(1)
		return
	}

	// custom setup

	//globalAccount := "5682a516-705b-4871-8480-54ed3caa2010"
	//subAccount := "60c5ea7a-7f2c-4ebd-9617-ebcc7310f8e5"
	runtimeID := "f86617c2-6c8d-417d-b980-6048aee020c7"

	//id, err := simu.Keb().CreateInstance(ctx, sim.CreateInstanceInput{
	//	Alias:         "shared-gcp",
	//	GlobalAccount: globalAccount,
	//	SubAccount:    subAccount,
	//	Provider:      cloudcontrolv1beta1.ProviderGCP,
	//})
	//if err != nil {
	//	logger.Error(err, "Failed to create instance")
	//	os.Exit(1)
	//}

	id, err := simu.Keb().GetInstance(ctx, runtimeID)
	if err != nil {
		logger.Error(err, "Failed to get instance")
		os.Exit(1)
	}

	fmt.Printf("Instance %q %q %q\n", id.Alias, id.ShootName, id.RuntimeID)

	// start sim

	go func() {
		if err := simu.Start(ctx); err != nil {
			logger.Error(err, "Failed running sim")
			os.Exit(1)
		}
	}()

	// wait

	logger.WithValues("shoot", id.ShootName).Info("Waiting for shoot to become ready")
	err = simu.Keb().WaitProvisioningCompleted(ctx, sim.WithRuntimes{id.RuntimeID}, sim.WithTimeout(10*time.Minute))
	if err != nil {
		logger.Error(err, "Failed to wait for instance to become ready")
	} else {
		logger.Info("Shoot is ready")
	}

	time.Sleep(15 * time.Minute)
}
