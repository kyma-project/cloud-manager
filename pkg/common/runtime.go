package common

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
)

const runtimeSecurityEnabledLabel = "cloud-manager.kyma-project.io/security-enabled"

func IsSecurityScanEnabledOnRuntime(obj *infrastructuremanagerv1.Runtime) bool {
	if composed.IsMarkedForDeletion(obj) {
		return false
	}
	_, ok := obj.Labels[runtimeSecurityEnabledLabel]
	return ok
}
