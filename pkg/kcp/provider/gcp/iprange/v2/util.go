package v2

import "fmt"

func GetIpRangeName(kcpResourceName string) string {
	return fmt.Sprintf("cm-%s", kcpResourceName)
}
