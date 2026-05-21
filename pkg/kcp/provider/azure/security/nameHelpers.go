package security

import (
	"fmt"
	"regexp"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

const (
	storageAccountPrefix        = "kymasec" // 7 chars
	maxStorageAccountNameLength = 24        // Azure constraint
)

func ResourceGroupDataName(shootName string) string {
	return fmt.Sprintf("kyma-security-%s", shootName)
}

func ResourceGroupWatcherName() string {
	return "NetworkWatcherRG"
}

func NetworkWatcherName(location string) string {
	return fmt.Sprintf("NetworkWatcher_%s", location)
}

func storageAccountBaseName(shootName string) string {
	name := strings.ToLower(shootName)
	name = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(name, "")
	maxShootPart := maxStorageAccountNameLength - len(storageAccountPrefix) - 5 // 5 chars for random suffix to ensure uniqueness
	if len(name) > maxShootPart {
		name = name[:maxShootPart]
	}
	return fmt.Sprintf("%s%s", storageAccountPrefix, name)
}

func StorageAccountNameAttempt(i int, shootName string) string {
	result := storageAccountBaseName(shootName)
	if i > 0 {
		// not first attempt, add random suffix
		suffixLen := maxStorageAccountNameLength - len(result)
		suffix := strings.ToLower(util.RandomString(suffixLen))
		result = fmt.Sprintf("%s%s", result, suffix)
	}
	return result
}

func FlowLogName(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) string {
	return vpcNetwork.Status.Identifiers.Name
}
