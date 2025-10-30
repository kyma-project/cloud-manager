package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func SkipE2eTests(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") == "" {
		t.Skip("Skipping E2E tests because $RUN_E2E_TESTS is not set")
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
