package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/e2e"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	var all bool
	var gcp bool
	var azure bool
	var aws bool
	var openstack bool
	var filename string

	flag.BoolVar(&all, "all", all, "Create SKRs for all providers")
	flag.BoolVar(&gcp, "gcp", gcp, "Create SKR for GCP provider")
	flag.BoolVar(&azure, "azure", azure, "Create SKR for Azure provider")
	flag.BoolVar(&aws, "aws", aws, "Create SKR for AWS provider")
	flag.BoolVar(&openstack, "openstack", openstack, "Create SKR for OpenStack provider")
	flag.StringVar(&filename, "filename", filename, "Output filename, defaults to ${PROJECTROOT}/.runtimes.yaml")
	flag.StringVar(&filename, "f", filename, "Output filename, defaults to ${PROJECTROOT}/.runtimes.yaml")

	flag.Parse()

	opts := zap.Options{}
	opts.Development = true
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	if !(all || gcp || azure || aws || openstack) {
		logger.Error(fmt.Errorf("usage error"), "At least one provider must be specified")
		os.Exit(1)
	}

	cfg := e2econfig.LoadConfig()

	if filename == "" {
		filename = path.Join(cfg.GetBaseDir(), ".runtimes.yaml")
	}
	_, err := e2e.LoadSharedState(filename)
	if err != nil && !os.IsNotExist(err) {
		logger.WithValues("path", filename).Error(err, "Failed to stat output file")
		os.Exit(3)
	}
	if err == nil {
		logger.Error(fmt.Errorf("output file already exists"), "Did you forgot to clean previously created SKRs?")
		os.Exit(4)
	}

	ctx := ctrl.SetupSignalHandler()
	ctx = e2e.NewScenarioSession(ctx)

	f := e2e.NewWorldFactory()
	world, err := f.Create(ctx)
	if err != nil {
		logger.Error(err, "Failed to create world")
		os.Exit(1)
	}

	var providers []cloudcontrolv1beta1.ProviderType
	if gcp || all {
		providers = append(providers, cloudcontrolv1beta1.ProviderGCP)
	}
	if azure || all {
		providers = append(providers, cloudcontrolv1beta1.ProviderAzure)
	}
	if aws || all {
		providers = append(providers, cloudcontrolv1beta1.ProviderAws)
	}
	if openstack {
		providers = append(providers, cloudcontrolv1beta1.ProviderOpenStack)
	}
	providers = pie.Unique(providers)

	var wg sync.WaitGroup
	var runtimes []string

	exitCode := 0

	for _, provider := range providers {
		wg.Add(1)
		go func(provider cloudcontrolv1beta1.ProviderType) {
			defer wg.Done()
			logger := logger.WithValues("provider", provider)
			subscription := e2e.Config.Subscriptions.GetDefaultForProvider(provider)
			if subscription == nil {
				logger.Error(fmt.Errorf("no subscription"), "Subscription for provider not configured")
				exitCode = 14
				return
			}
			alias := fmt.Sprintf("shared-%s", provider)
			logger = logger.WithValues("alias", alias)
			logger.Info("Creating SKR...")
			skr, _, err := world.SKR().CreateSkr(ctx, alias, e2e.CreatorSkrInput{
				Subscription: subscription,
				Logger:       logger,
			})
			if err != nil {
				logger.Error(err, "Failed to create SKR")
				exitCode = 10
				return
			}
			logger.
				WithValues(
					"runtime", skr.RuntimeID,
					"shoot", skr.ShootName,
				).Info("Created")
			runtimes = append(runtimes, skr.RuntimeID)
		}(provider)
	}

	wg.Wait()

	if err := world.Stop(ctx); err != nil {
		logger.Error(err, "Failed to stop world")
		exitCode = 11
	}

	if len(runtimes) > 0 {
		state := &e2e.SharedState{
			Runtimes: runtimes,
		}
		err = e2e.SaveSharedState(state, filename)
		if err != nil {
			logger.Error(err, "Failed to write output file")
			exitCode = 13
		}
	}

	os.Exit(exitCode)
}
