package client

import "fmt"

const (
	storageAccountPathPattern = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Storage/storageAccounts/%v"
	backupPolicyPathPattern   = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupPolicies/%v"
	vaultPathPattern          = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v"
	containerNamePattern      = "StorageContainer;Storage;%v;%v"
	recoveryPointPathPattern  = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupFabrics/Azure/protectionContainers/%v/protectedItems/%v/recoveryPoints/%v"
)

func GetStorageAccountPath(subscriptionId, resourceGroupName, storageAccountName string) string {
	return fmt.Sprintf(storageAccountPathPattern, subscriptionId, resourceGroupName, storageAccountName)
}

func GetBackupPolicyPath(subscriptionId, resourceGroupName, vaultName, backupPolicyName string) string {
	return fmt.Sprintf(backupPolicyPathPattern, subscriptionId, resourceGroupName, vaultName, backupPolicyName)
}

func GetVaultPath(subscriptionId, resourceGroupName, vaultName string) string {
	return fmt.Sprintf(vaultPathPattern, subscriptionId, resourceGroupName, vaultName)
}

func GetContainerName(resourceGroupName, storageAccountName string) string {
	return fmt.Sprintf(containerNamePattern, resourceGroupName, storageAccountName)
}

func GetRecoveryPointPath(subscriptionId, resourceGroupName, vaultName, storageAccountName, protectedItemName, recoveryPointName string) string {
	containerName := GetContainerName(resourceGroupName, storageAccountName)
	return fmt.Sprintf(recoveryPointPathPattern, subscriptionId, resourceGroupName, vaultName, containerName, protectedItemName, recoveryPointName)
}
