package util

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"strings"
)

func IsCrdDisabled(env abstractions.Environment, crd string) bool {
	disabledCrds := strings.Split(env.Get("DISABLED_CRDS"), ",")
	for _, disabled := range disabledCrds {
		if strings.ToLower(strings.TrimSpace(disabled)) == strings.ToLower(crd) {
			return true
		}
	}
	return false
}

func GetDisabledCrds(env abstractions.Environment) []string {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	disabledCrds := strings.Split(env.Get("DISABLED_CRDS"), ",")
	for i, disabled := range disabledCrds {
		disabledCrds[i] = strings.ToLower(strings.TrimSpace(disabled))
	}
	return disabledCrds
}
