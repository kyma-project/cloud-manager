package e2e

import (
	"fmt"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

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