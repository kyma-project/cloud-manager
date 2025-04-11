package subnet

import "fmt"

func GetSubnetShortName(objName string) string {
	return fmt.Sprintf("cm-%s", objName)
}

func GetSubnetFullName(projectId, region, subnetShortName string) string {
	return fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectId, region, subnetShortName)
}

func GetServiceConnectionPolicyShortName(network, region string) string {
	return fmt.Sprintf("cm-%s-%s-rc", network, region)
}

func GetServiceConnectionPolicyFullName(projectId, region, network string) string {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, region)
	connectionPolicyNameShort := GetServiceConnectionPolicyShortName(network, region)
	return fmt.Sprintf("%s/serviceConnectionPolicies/%s", parent, connectionPolicyNameShort)
}
