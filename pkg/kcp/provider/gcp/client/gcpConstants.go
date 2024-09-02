package client

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

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

const GcpRetryWaitTime = time.Second * 3
const GcpOperationWaitTime = time.Second * 5
const GcpApiTimeout = time.Second * 3

type GcpConfig struct {
	GcpRetryWaitTime     time.Duration
	GcpOperationWaitTime time.Duration
	GcpApiTimeout        time.Duration
}

func GetGcpConfig(env abstractions.Environment) *GcpConfig {
	return &GcpConfig{
		GcpRetryWaitTime:     GetConfigDuration(env, "GCP_RETRY_WAIT_DURATION", GcpRetryWaitTime),
		GcpOperationWaitTime: GetConfigDuration(env, "GCP_OPERATION_WAIT_DURATION", GcpOperationWaitTime),
		GcpApiTimeout:        GetConfigDuration(env, "GCP_API_TIMEOUT_DURATION", GcpApiTimeout),
	}
}

func GetConfigDuration(env abstractions.Environment, key string, defaultValue time.Duration) time.Duration {
	duration, err := time.ParseDuration(env.Get(key))
	if err != nil {
		return defaultValue
	}
	return duration
}

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

func GetCompleteServiceName(projectId string, serviceName GcpServiceName) string {
	return fmt.Sprintf("projects/%s/services/%s", projectId, serviceName)
}
