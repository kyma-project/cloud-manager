package e2e

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

const (
	sharedSkrId = "bcb334ad-9a19-4d67-aab2-ba1520fe8f21"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return setWorld(ctx, NewWorld()), nil
	})

	ctx.Step(`^there is SKR with "(AWS|Azure|GCP|OpenStack")" provider and default IpRange$`, thereIsSKRWithProviderAndDefaultIpRange)
	ctx.Step(`^(SKR|KCP|Garden) resource declaration:$`, resourceDeclaration)
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

func thereIsSKRWithProviderAndDefaultIpRange(ctx context.Context, provider string) (context.Context, error) {
	world := getWorld(ctx)
	pt, err := cloudcontrolv1beta1.ParseProviderType(provider)
	if err != nil {
		return ctx, err
	}

	sub := SubscriptionRegistry.GetDefaultForProvider(pt)
	if sub == nil {
		return ctx, fmt.Errorf("no default subscription found for provider %q", provider)
	}

	skrCreator := NewSkrCreator(world, sub)
	_, err = skrCreator.CreateSkr(ctx, pt)
	if err != nil {
		return ctx, fmt.Errorf("error creating SKR: %w", err)
	}
	return ctx, nil
}

func resourceDeclaration(ctx context.Context, plane string, tbl *godog.Table) (context.Context, error) {
	rd, err := ParseResourceDeclarations(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource declaration: %w", err)
	}

	world := getWorld(ctx)

	switch plane {
	case "SKR":
		return skrResourceDeclaration(ctx, sharedSkrId, tbl)
	case "KCP":
		kcp, err := world.ClusterProvider.KCP(ctx)
		if err != nil {
			return ctx, err
		}
		err = kcp.AddResources(ctx, rd)
		return ctx, err
	case "Garden":
		gc, err := world.ClusterProvider.Garden(ctx)
		if err != nil {
			return ctx, err
		}
		err = gc.AddResources(ctx, rd)
		return ctx, err
	default:
		return ctx, fmt.Errorf("unknown plane %q", plane)
	}
}

func skrResourceDeclaration(ctx context.Context, skrId string, tbl *godog.Table) (context.Context, error) {
	rd, err := ParseResourceDeclarations(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource declaration: %w", err)
	}

	world := getWorld(ctx)

	skr, err := world.ClusterProvider.SKR(ctx, skrId)
	if err != nil {
		return ctx, fmt.Errorf("failed to get SKR cluster: %w", err)
	}
	err = skr.AddResources(ctx, rd)
	return ctx, err
}

func resourceIsCreated(ctx context.Context, alias string, doc *godog.DocString) (context.Context, error) {
	return ctx, nil
}

func resourceIsDeleted(ctx context.Context, alias string) (context.Context, error) {
	return ctx, nil
}

func eventuallyValueIsOkUnless(ctx context.Context, expression string, unless *godog.Table) (context.Context, error) {
	return ctx, nil
}

func eventuallyValueIsOk(ctx context.Context, expression string) (context.Context, error) {
	return ctx, nil
}

func valueIsOk(ctx context.Context, expression string) (context.Context, error) {
	return ctx, nil
}

func pvcFileOperationsSucceed(ctx context.Context, alias string, ops *godog.Table) (context.Context, error) {
	return ctx, nil
}

func eventuallyResourceDoesNotExist(ctx context.Context, alias string) (context.Context, error) {
	return ctx, nil
}

func resourceDoesNotExist(ctx context.Context, alias string) (context.Context, error) {
	return ctx, nil
}

func logsOfContainerInPodContain(ctx context.Context, containerName string, alias string, content string) (context.Context, error) {
	return ctx, nil
}

func httpOperationSucceeds(ctx context.Context, ops *godog.Table) (context.Context, error) {
	return ctx, nil
}

func redisGivesWith(ctx context.Context, cmd string, out string, tbl *godog.Table) (context.Context, error) {
	return ctx, nil
}
