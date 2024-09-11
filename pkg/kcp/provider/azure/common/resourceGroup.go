package common

import "fmt"

// AzureCloudManagerResourceGroupName returns the common Cloud Manager resource group and
// virtual network name based on the gardener network name, ie shoot technicalId in the
// form of shoot--kyma-dev--c-123abc
func AzureCloudManagerResourceGroupName(gardenerNetworkName string) string {
	return fmt.Sprintf("%s-%s", gardenerNetworkName, "-cm")
}
