package common

import "fmt"

func KcpNetworkKymaCommonName(kymaName string) string {
	return fmt.Sprintf("%s--kyma", kymaName)
}

func KcpNetworkCMCommonName(kymaName string) string {
	return fmt.Sprintf("%s--cm", kymaName)
}
