package e2e

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func thereIsSharedSKRWithProvider(ctx context.Context, provider string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	alias := fmt.Sprintf("shared-%s", provider)
	_, err := session.AddExistingCluster(ctx, alias)
	return ctx, err
}

func moduleIsAdded(ctx context.Context, moduleName string) (context.Context, error) {
	return ctx, fmt.Errorf("not implemented")
}

func resourceDeclaration(ctx context.Context, tbl *godog.Table) (context.Context, error) {
	rd, err := ParseResourceDeclarations(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource declaration: %w", err)
	}

	if GetCurrentScenarioSession(ctx).CurrentCluster() == nil {
		return ctx, fmt.Errorf("current cluster is not defined")
	}

	err = GetCurrentScenarioSession(ctx).CurrentCluster().AddResources(ctx, rd...)
	if err != nil {
		return ctx, fmt.Errorf("error adding resource declaration: %w", err)
	}

	return ctx, nil
}

func resourceIsCreated(ctx context.Context, alias string, doc *godog.DocString) (context.Context, error) {
	arr, err := util.YamlMultiDecodeToUnstructured([]byte(doc.Content))
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource yaml: %w", err)
	}
	if len(arr) != 1 {
		return ctx, fmt.Errorf("expected one resource in yaml but got %d", len(arr))
	}
	obj := arr[0]

	err = GetCurrentScenarioSession(ctx).CurrentCluster().GetClient().Create(ctx, obj)

	return ctx, err
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
