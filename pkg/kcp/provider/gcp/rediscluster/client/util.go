package client

import (
	"fmt"
)

func GetGcpMemoryStoreRedisClusterName(projectId, locationId, instanceId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectId, locationId, GetGcpMemoryStoreRedisClusterId(instanceId))
}

func GetGcpMemoryStoreRedisClusterId(instanceId string) string {
	return fmt.Sprintf("cm-%s", instanceId)
}
