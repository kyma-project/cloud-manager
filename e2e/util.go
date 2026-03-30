package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

func SkipE2eTests(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") == "" {
		t.Skip("Skipping E2E tests because $RUN_E2E_TESTS is not set. This is OK if you're running normal unit tests for CI.")
	}
}

func SharedSkrClusterAlias(pt cloudcontrolv1beta1.ProviderType) string {
	return fmt.Sprintf("shared_%s", strings.ToLower(string(pt)))
}

func SharedRuntimeResourceAlias(runtimeId string) string {
	return fmt.Sprintf("rt_%s", runtimeId)
}

func SharedShootResourceAlias(shootName string) string {
	return fmt.Sprintf("shoot_%s", shootName)
}

func SharedGardenerClusterResourceAlias(name string) string {
	return fmt.Sprintf("gc_%s", name)
}

func waitClusterStarts(ctx context.Context, c cluster.Cluster) bool {
	var toCtx context.Context
	var toCancel context.CancelFunc
	if debugged.Debugged {
		toCtx, toCancel = context.WithTimeout(ctx, 10*time.Minute)
	} else {
		toCtx, toCancel = context.WithTimeout(ctx, 10*time.Second)
	}
	defer toCancel()
	if ok := c.GetCache().WaitForCacheSync(toCtx); !ok {
		return false
	}
	return true
}
