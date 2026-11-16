package util

import "fmt"

func WafLogGroupName(suffix string) string {
	return fmt.Sprintf("aws-waf-logs-%s", suffix)
}
