package redisinstance

import (
	"fmt"
	"strings"
)

func GetAwsElastiCacheSubnetGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheParameterGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheParameterGroupFamily(engineVersion string) string {
	if strings.Contains(engineVersion, "2.6") {
		return "redis2.6"
	}
	if strings.Contains(engineVersion, "2.8") {
		return "redis2.8"
	}
	if strings.Contains(engineVersion, "3.2") {
		return "redis3.2"
	}
	if strings.Contains(engineVersion, "4.0") {
		return "redis4.0"
	}
	if strings.Contains(engineVersion, "5.0") {
		return "redis5.0"
	}
	if strings.Contains(engineVersion, "6.") {
		return "redis6.x"
	}
	if strings.Contains(engineVersion, "7.") {
		return "redis7"
	}

	return ""
}

func GetAwsElastiCacheClusterName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}
