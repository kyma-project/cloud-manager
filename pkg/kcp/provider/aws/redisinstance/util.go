package redisinstance

import "fmt"

func GetAwsElastiCacheSubnetGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheClusterName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}
