package common

import (
	"fmt"
	"strings"

	vpcnetworkconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/config"
)

func GardenerVpcName(shootNamespace, shootName string) string {
	project := strings.TrimPrefix(shootNamespace, "garden-")
	return fmt.Sprintf("shoot--%s--%s", project, shootName)
}

func KymaVpcName(kcpVpcObjName string) string {
	var result string
	if vpcnetworkconfig.VpcNetworkConfig.Prefix != "" {
		result = fmt.Sprintf("kyma-%s-%s", vpcnetworkconfig.VpcNetworkConfig.Prefix, kcpVpcObjName)
	} else {
		result = fmt.Sprintf("kyma-%s", kcpVpcObjName)
	}
	if len(result) > 60 {
		result = result[:60]
	}
	return result
}
