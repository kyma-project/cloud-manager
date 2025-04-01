package common

import (
	"fmt"
	"strings"
)

func GardenerVpcName(shootNamespace, shootName string) string {
	project := strings.TrimPrefix(shootNamespace, "garden-")
	return fmt.Sprintf("shoot--%s--%s", project, shootName)
}
