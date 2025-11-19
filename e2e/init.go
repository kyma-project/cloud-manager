package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func InitializeTestSuite(gdCtx *godog.TestSuiteContext) {
	gdCtx.BeforeSuite(func() {
		opts := zap.Options{}
		opts.Development = true
		logger := zap.New(zap.UseFlagOptions(&opts))
		ctrl.SetLogger(logger)

		config := e2econfig.LoadConfig()

		ctx := context.Background()

		f := NewWorldFactory()
		w, err := f.Create(ctx, WorldCreateOptions{
			Config: config,
		})
		if err != nil {
			panic(err)
		}
		world = w
		time.Sleep(10 * time.Second)
	})

	gdCtx.AfterSuite(func() {
		if world != nil {
			world.Cancel()
			world.StopWaitGroup().Wait()
			if world.RunError() != nil {
				panic(world.RunError())
			}
		}
	})
}

var world WorldIntf

func GetWorld() WorldIntf {
	return world
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if GetWorld() == nil {
			return ctx, fmt.Errorf("world does not exist")
		}

		ctx = StartNewScenarioSession(ctx)

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		var result error
		if err != nil {
			result = multierror.Append(result, err)
		}
		session := GetCurrentScenarioSession(ctx)
		if session != nil {
			if err := session.Terminate(ctx); err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to terminate session: %w", err))
			}
		}

		return ctx, result
	})

	ctx.Step(`^there is shared SKR with "(AWS|Azure|GCP|OpenStack")" provider$`, thereIsSharedSKRWithProvider)

	ctx.Step(`^module "([^"]*)" is active`, moduleIsActive)
	ctx.Step(`^module "([^"]*)" is active`, moduleIsNotActive)
	ctx.Step(`^module "([^"]*)" is added$`, moduleIsAdded)
	ctx.Step(`^module "([^"]*)" is removed`, moduleIsRemoved)

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
