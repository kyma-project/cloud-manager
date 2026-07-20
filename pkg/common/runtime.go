package common

import (
	"strconv"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
)

const TmpRuntimeSecurityEnabledLabel = "cloud-manager.kyma-project.io/security-enabled"

func IsSecurityScanEnabledOnRuntime(obj *infrastructuremanagerv1.Runtime) bool {
	if composed.IsMarkedForDeletion(obj) {
		return false
	}
	val, ok := obj.Labels[TmpRuntimeSecurityEnabledLabel]
	if !ok {
		return false
	}
	ok, _ = strconv.ParseBool(val)
	return ok
}
