package main

import (
	"fmt"
	"os"
)

func mustAll(errList ...error) {
	for _, err := range errList {
		if err != nil {
			panic(err)
		}
	}
}

func requiredString(val, name string) error {
	if len(val) == 0 {
		return fmt.Errorf("argument %s is required", name)
	}
	return nil
}

func requiredKymaName() error {
	return requiredString(kymaName, "kymaName")
}

func requiredKymaModuleName() error {
	return requiredString(kymaModuleName, "module")
}

func requiredKymaModuleState() error {
	return requiredString(moduleState, "state")
}

func requiredShootName() error {
	return requiredString(shootName, "shoot")
}

func defaultKcpNamespace() error {
	if namespace == "" {
		namespace = os.Getenv("CM_NAMESPACE_KCP")
	}
	if namespace == "" {
		namespace = "kcp-system"
	}
	return nil
}

func defaultGardenNamespace() error {
	if namespace == "" {
		namespace = os.Getenv("CM_NAMESPACE_GARDEN")
	}
	if namespace == "" {
		namespace = "garden-kyma"
	}
	return nil
}
