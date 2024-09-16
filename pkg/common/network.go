package common

import "fmt"

func KymaNetworkCommonName(kymaName string) string {
	return fmt.Sprintf("%s-kyma", kymaName)
}
