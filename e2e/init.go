package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
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
			Config:      config,
			CreateCloud: true,
		})
		if err != nil {
			panic(err)
		}
		world = w
		time.Sleep(1 * time.Second)
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
	ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		GetCurrentScenarioSession(ctx).SetStepName(st.Text)
		return ctx, nil
	})
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if GetWorld() == nil {
			return ctx, fmt.Errorf("world does not exist")
		}

		ctx = StartNewScenarioSession(ctx, sc.Name)

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		session := GetCurrentScenarioSession(ctx)
		if session != nil {
			if err := session.Terminate(ctx); err != nil {
				return ctx, fmt.Errorf("failed to terminate session: %w", err)
			}
		}

		return ctx, nil
	})

	ctx.Step(`^debug wait "([^"]*)"$`, debugWait)
	ctx.Step(`^debug log (true|false)$`, debugLog)

	ctx.Step(`^eventually timeout is "([^"]*)"$`, eventuallyTimeoutIs)

	ctx.Step(`^there is shared SKR with "(AWS|Azure|GCP|OpenStack)" provider$`, thereIsSharedSKRWithProvider)

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
	ctx.Step(`^tf module "([^"]*)" is applied:$`, tfModuleIsApplied)
	ctx.Step(`^tf module "([^"]*)" is destroyed$`, tfModuleIsDestroyed)
	ctx.Step(`^current cluster is "([^"]*)"$`, currentClusterIs)
	ctx.Step(`^Subscription "([^"]+)" exists for "(AWS|Azure|GCP|OpenStack)" provider$`, subscriptionExistsForProvider)
}
