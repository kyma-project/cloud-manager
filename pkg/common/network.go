package common

import (
	"fmt"
)

func KcpNetworkKymaCommonName(kymaName string) string {
	return fmt.Sprintf("%s--kyma", kymaName)
}

func KcpNetworkCMCommonName(kymaName string) string {
	return fmt.Sprintf("%s--cm", kymaName)
}

func IsKcpNetworkKyma(kcpNetworkObjName, kymaOrScopeName string) bool {
	return kcpNetworkObjName == KcpNetworkKymaCommonName(kymaOrScopeName)
}

func IsKcpNetworkCM(kcpNetworkObjName, kymaOrScopeName string) bool {
	return kcpNetworkObjName == KcpNetworkCMCommonName(kymaOrScopeName)
}
