package client

import (
	"fmt"
	"regexp"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/util"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

const vPCPathPattern = "projects/%s/global/networks/%s"
const ServiceNetworkingServicePath = "services/servicenetworking.googleapis.com"
const ServiceNetworkingServiceConnectionName = "services/servicenetworking.googleapis.com/connections/servicenetworking-googleapis-com"
const networkFilter = "network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\""
const PsaPeeringName = "servicenetworking-googleapis-com"
const filestoreInstancePattern = "projects/%s/locations/%s/instances/%s"
const filestoreParentPattern = "projects/%s/locations/%s"
const fileBackupPattern = "projects/%s/locations/%s/backups/%s"
const fileBackupRegex = `projects/(?P<Project>[^/]+)/locations/(?P<Location>[^/]+)/backups/(?P<BackupName>[^/]+)`
const fileBackupUrlRegex = `locations/(?P<Location>[^/]+)/backups/(?P<BackupName>[^/]+)`

const GcpRetryWaitTime = time.Second * 3
const GcpOperationWaitTime = time.Second * 5
const GcpCapacityCheckInterval = time.Hour * 1

const skrBackupsFilter = "labels.managed-by=\"%s\" AND labels.scope-name=\"%s\""

const sharedBackupsFilter = "labels.managed-by=\"%s\" AND " +
	"( labels.cm-allow-%s=\"" + util.GcpLabelBackupAccessibleFrom + "\"" +
	" OR labels.cm-allow-%s=\"" + util.GcpLabelBackupAccessibleFrom + "\"" +
	" OR labels.ALL=\"" + util.GcpLabelBackupAccessibleFrom + "\")"

const GcpNfsStateDataProtocol = "gcpNfsProtocol"

type FilestoreProtocol string

const (
	FilestoreProtocolNFSv41 FilestoreProtocol = "NFS_V4_1"
)

func GetVPCPath(projectId, vpcId string) string {
	return fmt.Sprintf(vPCPathPattern, projectId, vpcId)
}

func GetNetworkFilter(projectId, vpcId string) string {
	return fmt.Sprintf(networkFilter, projectId, vpcId)
}

func GetSkrBackupsFilter(scopeName string) string {
	return fmt.Sprintf(skrBackupsFilter, ManagedByValue, scopeName)
}

func GetSharedBackupsFilter(shootName, subaccountId string) string {
	return fmt.Sprintf(sharedBackupsFilter, ManagedByValue, shootName, subaccountId)
}

func GetFilestoreInstancePath(projectId, location, instanceId string) string {
	return fmt.Sprintf(filestoreInstancePattern, projectId, location, instanceId)
}

func GetFilestoreParentPath(projectId, location string) string {
	return fmt.Sprintf(filestoreParentPattern, projectId, location)
}

func GetFileBackupPath(projectId, location, name string) string {
	return fmt.Sprintf(fileBackupPattern, projectId, location, name)
}

type addressType string

const (
	AddressTypeInternal addressType = "INTERNAL"
)

type ipRangePurpose string

const (
	IpRangePurposeVPCPeering ipRangePurpose = "VPC_PEERING"
)

const (
	//Common States
	Deleted  v1beta1.StatusState = "Deleted"
	Creating v1beta1.StatusState = "Creating"
	Updating v1beta1.StatusState = "Updating"
	Deleting v1beta1.StatusState = "Deleting"

	//IPRange States
	SyncAddress         v1beta1.StatusState = "SyncAddress"
	SyncPsaConnection   v1beta1.StatusState = "SyncPSAConnection"
	DeletePsaConnection v1beta1.StatusState = "DeletePSAConnection"
	DeleteAddress       v1beta1.StatusState = "DeleteAddress"
)

type FilestoreState string

const (
	READY    FilestoreState = "READY"
	DELETING FilestoreState = "DELETING"
	ERROR    FilestoreState = "ERROR"
)

type OperationType int

const (
	NONE OperationType = iota
	ADD
	MODIFY
	DELETE
)

type GcpServiceName string

const (
	ServiceNetworkingService    GcpServiceName = "servicenetworking.googleapis.com"
	ComputeService              GcpServiceName = "compute.googleapis.com"
	FilestoreService            GcpServiceName = "file.googleapis.com"
	CloudResourceManagerService GcpServiceName = "cloudresourcemanager.googleapis.com"
	MemoryStoreForRedisService  GcpServiceName = "redis.googleapis.com"
)

// Used for Hyperscaler labels mainly backups which need to be nuked without KCP mirror object
const (
	ManagedByKey   = "managed-by"
	ManagedByValue = "cloud-manager"
	ScopeNameKey   = "scope-name"
)

func GetCompleteServiceName(projectId string, serviceName GcpServiceName) string {
	return fmt.Sprintf("projects/%s/services/%s", projectId, serviceName)
}

func GetProjectLocationNameFromFileBackupPath(fullPath string) (string, string, string) {
	re := regexp.MustCompile(fileBackupRegex)
	matches := re.FindStringSubmatch(fullPath)
	return matches[re.SubexpIndex("Project")], matches[re.SubexpIndex("Location")], matches[re.SubexpIndex("BackupName")]
}
func GetLocationNameFromFileBackupUrl(backupUrl string) (string, string) {
	re := regexp.MustCompile(fileBackupUrlRegex)
	matches := re.FindStringSubmatch(backupUrl)
	return matches[re.SubexpIndex("Location")], matches[re.SubexpIndex("BackupName")]
}
