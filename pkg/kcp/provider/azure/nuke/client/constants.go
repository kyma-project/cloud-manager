package client

import (
	"fmt"
	"regexp"
)

const (
	fabricName              = "Azure"
	tagNameCloudManager     = "cloud-manager"
	tagValueRwxVolumeBackup = "rwxVolumeBackup"
)

const vaultIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)"
const protectedItemIdPattern = "\\/subscriptions\\/(?<subscription>[^\\/]*)\\/resourceGroups\\/(?<resourceGroup>[^\\/]*)\\/providers\\/Microsoft.RecoveryServices\\/vaults\\/(?<vault>[^\\/]*)\\/backupFabrics\\/Azure\\/protectionContainers\\/(?<container>[^\\/]*)\\/protectedItems\\/(?<protectedItem>[^\\/]*)"

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
