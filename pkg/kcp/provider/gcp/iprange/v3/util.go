package v3

import "fmt"

func GetPrivateSubnetShortName(objName string) string {
	return fmt.Sprintf("cm-%s", objName)
}

func GetServiceConnectionPolicyShortName(network, region string) string {
	return fmt.Sprintf("cm-%s-%s-redis-cluster", network, region)
}

func GetServiceConnectionPolicyFullName(projectId, region, network string) string {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, region)
	connectionPolicyNameShort := GetServiceConnectionPolicyShortName(network, region)
	return fmt.Sprintf("%s/serviceConnectionPolicies/%s", parent, connectionPolicyNameShort)
}
