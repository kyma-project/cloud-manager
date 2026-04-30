package common

import "github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"

const securityEnabledLabel = "cloud-manager.kyma-project.io/tmp-security-enabled"

func IsSecurityScanEnabledOnRuntime(obj *infrastructuremanagerv1.Runtime) bool {
	_, ok := obj.Labels[securityEnabledLabel]
	return ok
}
