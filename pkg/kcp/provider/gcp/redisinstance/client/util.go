package client

import "fmt"

func GetGcpMemoryStoreRedisName(projectId, locationId, instanceId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, locationId, GetGcpMemoryStoreRedisInstanceId(instanceId))
}

func GetGcpMemoryStoreRedisInstanceId(instanceId string) string {
	return fmt.Sprintf("cm-%s", instanceId)
}
