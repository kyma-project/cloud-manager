package client

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
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
const vaultIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)"
const protectedItemIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)\\/backupFabrics\\/Azure\\/protectionContainers\\/(?<container>[^\\/]*)\\/protectedItems\\/AzureFileShare;(?<protectedItem>[^\\/]*)"
const containerIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)\\/backupFabrics\\/Azure\\/protectionContainers\\/(?<container>[^\\/]*)"

const (
	storageAccountPathPattern = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Storage/storageAccounts/%v"
	backupPolicyPathPattern   = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupPolicies/%v"
	vaultPathPattern          = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v"
	containerNamePattern      = "StorageContainer;Storage;%v;%v"
	fileShareNamePattern      = "AzureFileShare;%v"
	recoveryPointPathPattern  = "/subscriptions/%v/resourceGroups/%v/providers/Microsoft.RecoveryServices/vaults/%v/backupFabrics/Azure/protectionContainers/%v/protectedItems/%v/recoveryPoints/%v"
	fileSharePathPattern      = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.RecoveryServices/vaults/%s/backupFabrics/Azure/protectionContainers/%s/protectedItems/AzureFileShare;%s"
)

const storageProvisionerKey = "volume.kubernetes.io/storage-provisioner"
const azureFileShareProvisioner = "file.csi.azure.com"

const AzureFabricName = "Azure"
const TagNameCloudManager = "cloud-manager"
const TagValueRwxVolumeBackup = "rwxVolumeBackup"

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

func GetFileShareName(name string) string {
	return fmt.Sprintf(fileShareNamePattern, name)
}

func GetRecoveryPointPath(subscriptionId, resourceGroupName, vaultName, storageAccountName, protectedItemName, recoveryPointName string) string {
	containerName := GetContainerName(resourceGroupName, storageAccountName)
	return fmt.Sprintf(recoveryPointPathPattern, subscriptionId, resourceGroupName, vaultName, containerName, protectedItemName, recoveryPointName)
}

func GetFileSharePath(subscriptionId, resourceGroupName, vaultName, containerName, fileShareName string) string {
	return fmt.Sprintf(fileSharePathPattern, subscriptionId, resourceGroupName, vaultName, containerName, fileShareName)
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

func AzureStorageErrorInfoToJson(details []*armrecoveryservicesbackup.AzureStorageErrorInfo) (string, error) {
	jsonData, err := json.Marshal(details)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func ParseVaultId(vaultId string) (subscription string, resourceGroup string, vault string, err error) {
	re := regexp.MustCompile(vaultIdPattern)
	match := re.FindStringSubmatch(vaultId)
	if match == nil {
		return "", "", "", fmt.Errorf("vaultId %s does not match pattern %s", vaultId, vaultIdPattern)
	}
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["subscription"], result["resourceGroup"], result["vault"], nil
}

func ParseProtectedItemId(protectedId string) (subscription string, resourceGroup string, vault string, container string, protectedItem string, err error) {
	re := regexp.MustCompile(protectedItemIdPattern)
	match := re.FindStringSubmatch(protectedId)
	if match == nil {
		return "", "", "", "", "", fmt.Errorf("protectedItemId %s does not match pattern %s", protectedId, protectedItemIdPattern)
	}
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["subscription"], result["resourceGroup"], result["vault"], result["container"], result["protectedItem"], nil
}

func ParseContainerId(containerId string) (subscription string, resourceGroup string, vault string, container string, err error) {
	re := regexp.MustCompile(containerIdPattern)
	match := re.FindStringSubmatch(containerId)
	if match == nil {
		return "", "", "", "", fmt.Errorf("container id %s does not match pattern %s", containerId, containerIdPattern)
	}
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["subscription"], result["resourceGroup"], result["vault"], result["container"], nil
}

func ToStorageJobTimeFilter(time time.Time) string {
	return time.UTC().Format("2006-01-02 03:04:05 PM")
}
