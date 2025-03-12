package client

import (
	"fmt"
	"regexp"
)

type pvVolumeHandleRegexGroupKey string

const (
	vhStorageAccountName pvVolumeHandleRegexGroupKey = "storageAccountName"
	vhResourceGroupName  pvVolumeHandleRegexGroupKey = "resourceGroupName"
	vhFileShareName      pvVolumeHandleRegexGroupKey = "fileShareName"
	vhUuid               pvVolumeHandleRegexGroupKey = "uuid"
	vhSecretNamespace    pvVolumeHandleRegexGroupKey = "secretNamespace"
)

type recoverPointIdRegexGroupKey string

const (
	rpSubscription    recoverPointIdRegexGroupKey = "subscription"
	rpResourceGroup   recoverPointIdRegexGroupKey = "resourceGroup"
	rpVault           recoverPointIdRegexGroupKey = "vault"
	rpContainer       recoverPointIdRegexGroupKey = "container"
	rpProtectedItem   recoverPointIdRegexGroupKey = "protectedItem"
	rpRecoveryPointId recoverPointIdRegexGroupKey = "recoveryPointId"
)

const pvVolumeHandlePattern = "(?<resourceGroupName>[^\\#]*)#(?<storageAccountName>[^\\#]*)#(?<fileShareName>[^\\#]*)#(?<placeHolder>[^\\#]*)#(?<uuid>[^\\#]*)#(?<secretNamespace>[^\\#]+)"

const recoverPointIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)\\/backupFabrics\\/Azure\\/protectionContainers\\/(?<container>[^\\/]*)\\/protectedItems\\/(?<protectedItem>[^\\/]*)\\/recoveryPoints\\/(?<recoveryPointId>[^\\/]*)"

const (
	storageAccountPathPattern = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Storage/storageAccounts/%v"
	backupPolicyPathPattern   = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupPolicies/%v"
	vaultPathPattern          = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v"
	containerNamePattern      = "StorageContainer;Storage;%v;%v"
	recoveryPointPathPattern  = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupFabrics/Azure/protectionContainers/%v/protectedItems/%v/recoveryPoints/%v"
)

const storageProvisionerKey = "volume.kubernetes.io/storage-provisioner"
const azureFileShareProvisioner = "file.csi.azure.com"

const AzureFabricName = "Azure"

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
func ParsePvVolumeHandle(pvVolumeHandle string) (resourceGroupName string, storageAccountName string, fileShareName string, uuid string, secretNamespace string, err error) {
	re := regexp.MustCompile(pvVolumeHandlePattern)
	match := re.FindStringSubmatch(pvVolumeHandle)
	if match == nil {
		return "", "", "", "", "", fmt.Errorf("pvVolumeHandle %s does not match pattern %s", pvVolumeHandle, pvVolumeHandlePattern)
	}
	result := make(map[pvVolumeHandleRegexGroupKey]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[pvVolumeHandleRegexGroupKey(name)] = match[i]
		}
	}
	return result[vhResourceGroupName], result[vhStorageAccountName], result[vhFileShareName], result[vhUuid], result[vhSecretNamespace], nil
}

func ParseRecoveryPointId(recoveryPointId string) (subscription string, resourceGroup string, vault string, container string, protectedItem string, recoveryPointIdValue string, err error) {
	re := regexp.MustCompile(recoverPointIdPattern)
	match := re.FindStringSubmatch(recoveryPointId)
	if match == nil {
		return "", "", "", "", "", "", fmt.Errorf("recoveryPointId %s does not match pattern %s", recoveryPointId, recoverPointIdPattern)
	}
	result := make(map[recoverPointIdRegexGroupKey]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[recoverPointIdRegexGroupKey(name)] = match[i]
		}
	}
	return result[rpSubscription], result[rpResourceGroup], result[rpVault], result[rpContainer], result[rpProtectedItem], result[rpRecoveryPointId], nil
}

func IsPvcProvisionerAzureCsiDriver(annotations map[string]string) bool {
	provisioner, ok := annotations[storageProvisionerKey]
	return ok && provisioner == azureFileShareProvisioner
}
