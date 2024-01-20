package client

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
)

const vPCPathPattern = "projects/%s/global/networks/%s"
const ServiceNetworkingServicePath = "services/servicenetworking.googleapis.com"
const ServiceNetworkingServiceConnectionName = "services/servicenetworking.googleapis.com/connections/servicenetworking-googleapis-com"
const networkFilter = "network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\""
const PsaPeeringName = "servicenetworking-googleapis-com"
const filestoreInstancePattern = "projects/%s/locations/%s/instances/%s"
const filestoreParentPattern = "projects/%s/locations/%s"

func GetVPCPath(projectId, vpcId string) string {
	return fmt.Sprintf(vPCPathPattern, projectId, vpcId)
}

func GetNetworkFilter(projectId, vpcId string) string {
	return fmt.Sprintf(networkFilter, projectId, vpcId)
}

func GetFilestoreInstancePath(projectId, location, instanceId string) string {
	return fmt.Sprintf(filestoreInstancePattern, projectId, location, instanceId)
}

func GetFilestoreParentPath(projectId, location string) string {
	return fmt.Sprintf(filestoreParentPattern, projectId, location)
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

const (
	//Common States
	Deleted v1beta1.StatusState = "Deleted"

	//IPRange States
	SyncAddress         v1beta1.StatusState = "SyncAddress"
	SyncPsaConnection   v1beta1.StatusState = "SyncPSAConnection"
	DeletePsaConnection v1beta1.StatusState = "DeletePSAConnection"
	DeleteAddress       v1beta1.StatusState = "DeleteAddress"

	//Filestore States
	SyncFilestore   v1beta1.StatusState = "SyncFilestore"
	DeleteFilestore v1beta1.StatusState = "DeleteFilestore"
)

type FilestoreState string

const (
	CREATING   FilestoreState = "CREATING"
	READY      FilestoreState = "READY"
	REPAIRING  FilestoreState = "REPAIRING"
	DELETING   FilestoreState = "DELETING"
	ERROR      FilestoreState = "ERROR"
	RESTORING  FilestoreState = "RESTORING"
	SUSPENDED  FilestoreState = "SUSPENDED"
	SUSPENDING FilestoreState = "SUSPENDING"
	RESUMING   FilestoreState = "RESUMING"
	REVERTING  FilestoreState = "REVERTING"
)
