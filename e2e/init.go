package e2e

import (
	"context"
	"fmt"
	"path"

	"github.com/cucumber/godog"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		opts := zap.Options{}
		opts.Development = true
		logger := zap.New(zap.UseFlagOptions(&opts))
		ctrl.SetLogger(logger)

		cfg := e2econfig.LoadConfig()
		sharedStateFile := path.Join(cfg.GetBaseDir(), ".runtimes.yaml")
		sharedState, err := LoadSharedState(sharedStateFile)
		if err != nil {
			panic(fmt.Errorf("failed loading shared runtimes state: %w", err))
		}

		f := NewWorldFactory()
		w, err := f.Create(context.Background())
		if err != nil {
			panic(err)
		}
		world = w

		ctx := context.Background()
		for _, runtimeId := range sharedState.Runtimes {
			logger.WithValues("runtimeID", runtimeId).Info("Importing runtime...")
			skr, err := w.SKR().ImportShared(ctx, runtimeId)
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
	})

	ctx.AfterSuite(func() {
		if world != nil {
			err := world.Stop(context.Background())
			if err != nil {
				panic(err)
			}
		}
	})
}

var world World

func GetWorld() World {
	return world
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if GetWorld() == nil {
			return ctx, fmt.Errorf("world does not exist")
		}

		ctx = NewScenarioSession(ctx)

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		var result error
		if err != nil {
			result = multierror.Append(result, err)
		}
		if GetWorld() == nil {
			result = multierror.Append(result, fmt.Errorf("world does not exist"))
			return ctx, result
		}

		for _, alias := range GetScenarioSession(ctx).AllRegisteredClusters() {
			err = GetWorld().DeleteSKR(ctx, GetWorld().SKR().Get(alias))
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to stop transient SKR %q: %w", alias, err))
			}
		}

		return ctx, result
	})

	ctx.Step(`^there is SKR with "(AWS|Azure|GCP|OpenStack")" provider and default IpRange$`, thereIsSKRWithProviderAndDefaultIpRange)

	ctx.Step(`^module "([^"]*)" is added$`, moduleIsAdded)

	ctx.Step(`^resource declaration:$`, resourceDeclaration)
	ctx.Step(`^SKR "([^"]*)" resource declaration:$`, resourceDeclaration)
	ctx.Step(`^eventually "([^"]*)" is ok, unless:$`, eventuallyValueIsOkUnless)
	ctx.Step(`^"([^"]*)" is ok$`, valueIsOk)
	ctx.Step(`^eventually "([^"]*)" is ok$`, eventuallyValueIsOk)
	ctx.Step(`^PVC "([^"]*)" file operations succeed:$`, pvcFileOperationsSucceed)
	ctx.Step(`^resource "([^"]*)" is created:$`, resourceIsCreated)
	ctx.Step(`^resource "([^"]*)" is deleted$`, resourceIsDeleted)
	ctx.Step(`^eventually resource "([^"]*)" does not exist$`, eventuallyResourceDoesNotExist)
	ctx.Step(`^resource "([^"]*)" does not exist$`, resourceDoesNotExist)
	ctx.Step(`^logs of container "([^"]*)" in pod "([^"]*)" contain "([^"]*)"$`, logsOfContainerInPodContain)
	ctx.Step(`^HTTP operation succeeds:$`, httpOperationSucceeds)
	ctx.Step(`^Redis "([^"]*)" gives "([^"]*)" with:$`, redisGivesWith)
}
