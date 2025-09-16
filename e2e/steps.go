package e2e

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

func thereIsSKRWithProviderAndDefaultIpRange(ctx context.Context, provider string) (context.Context, error) {
	return ctx, fmt.Errorf("not implemented")
	//world := GetWorld()
	//pt, err := cloudcontrolv1beta1.ParseProviderType(provider)
	//if err != nil {
	//	return ctx, err
	//}

	//clusterAlias := SharedSkrClusterAlias(pt)
	//skr := world.SKR().GetByAlias(clusterAlias)
	//if skr == nil {
	//	return ctx, fmt.Errorf("could not find precreated cluster %q", clusterAlias)
	//}
	//
	//GetScenarioSession(ctx).SetCurrentCluster(skr, skr.Alias())
	//
	//return ctx, nil
}

func moduleIsAdded(ctx context.Context, moduleName string) (context.Context, error) {
	return ctx, fmt.Errorf("not implemented")
	//session, err := GetScenarioSessionEnsureCluster(ctx)
	//if err != nil {
	//	return ctx, err
	//}

	// TODO: continue here
	//session.CurrentCluster().

	return ctx, nil
}

func resourceDeclaration(ctx context.Context, tbl *godog.Table) (context.Context, error) {
	rd, err := ParseResourceDeclarations(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource declaration: %w", err)
	}

	if GetScenarioSession(ctx).CurrentCluster() == nil {
		return ctx, fmt.Errorf("current cluster is not defined")
	}

	err = GetScenarioSession(ctx).CurrentCluster().AddResources(ctx, rd...)
	if err != nil {
		return ctx, fmt.Errorf("error adding resource declaration: %w", err)
	}

	return ctx, nil
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
