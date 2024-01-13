package client

import (
	"fmt"
)

const vPCPathPattern = "projects/%s/global/networks/%s"
const ServiceNetworkingServicePath = "services/servicenetworking.googleapis.com"
const ServiceNetworkingServiceConnectionName = "services/servicenetworking.googleapis.com/connections/servicenetworking-googleapis-com"
const NetworkFilter = "network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\""

const PsaPeeringName = "servicenetworking-googleapis-com"

func GetVPCPath(projectId, vpcId string) string {
	return fmt.Sprintf(vPCPathPattern, projectId, vpcId)
}

type networkTier string

const (
	NetworkTierPremium  networkTier = "PREMIUM"
	NetworkTierStandard networkTier = "STANDARD"
)

type ipVersion string

const (
	IpVersionIpV4 ipVersion = "IPV4"
	IpVersionIpV6 ipVersion = "IPV6"
)

type addressType string

const (
	AddressTypeExternal addressType = "EXTERNAL"
	AddressTypeInternal addressType = "INTERNAL"
)

type ipRangePurpose string

const (
	IpRangePurposeVPCPeering            ipRangePurpose = "VPC_PEERING"
	IpRangePurposePrivateServiceConnect ipRangePurpose = "PRIVATE_SERVICE_CONNECT"
)

type ipv6EndpointType string

const (
	Ipv6EndpointTypeVm    ipv6EndpointType = "VM"
	Ipv6EndpointTypeNetlb ipv6EndpointType = "NETLB"
)
