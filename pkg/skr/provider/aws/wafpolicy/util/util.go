package util

import "fmt"

func WafLogGroupName(shootName string) string {
	return fmt.Sprintf("aws-waf-logs-%s", shootName)
}
