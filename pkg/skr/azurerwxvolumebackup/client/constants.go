package client

import "fmt"

const storageAccountPathPattern = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Storage/storageAccounts/%v"

func GetStorageAccountPath(subscriptionId, resourceGroupName, storageAccountName string) string {
	return fmt.Sprintf(storageAccountPathPattern, subscriptionId, resourceGroupName, storageAccountName)
}
