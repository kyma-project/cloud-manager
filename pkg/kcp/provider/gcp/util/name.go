package util

import (
	"fmt"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

type NameDetail struct {
	parts        []namePart
	resourceType ResourceType
}

// new name from pb objects -------------------------------------------------

func NewNameFromComputeOperation(op *computepb.Operation) NameDetail {
	return MustParseNameDetail(ptr.Deref(op.SelfLink, ""))
}

func NewNameFromOperation(op *longrunningpb.Operation) NameDetail {
	// TODO: ??? verify it
	return MustParseNameDetail(op.Name)
}

func NewNameFromNetwork(network *computepb.Network) NameDetail {
	return MustParseNameDetail(ptr.Deref(network.SelfLink, ""))
}

func NewNameFromSubnetwork(subnetwork *computepb.Subnetwork) NameDetail {
	return MustParseNameDetail(ptr.Deref(subnetwork.SelfLink, ""))
}

func NewNameFromAddress(address *computepb.Address) NameDetail {
	return MustParseNameDetail(ptr.Deref(address.SelfLink, ""))
}

func NewNameFromFilestoreInstance(filestore *computepb.Instance) NameDetail {
	return MustParseNameDetail(ptr.Deref(filestore.SelfLink, ""))
}

func NewNameFromRedisInstance(redis *redispb.Instance) NameDetail {
	return MustParseNameDetail(redis.Name)
}

func NewNameFromRedisCluster(rc *clusterpb.Cluster) NameDetail {
	return MustParseNameDetail(rc.Name)
}

// new name from primitives ---------------------------------------------

func NewGlobalOperationName(projectId, operationId string) NameDetail {
	return newNameDetail(ResourceTypeGlobalNetwork, nameDefnProject.Value(projectId), nameDefnOperation.Value(operationId))
}

func NewGlobalNetworkName(projectId, networkId string) NameDetail {
	return newNameDetail(ResourceTypeGlobalNetwork, nameDefnProject.Value(projectId), nameDefnGlobalNetwork.Value(networkId))
}

func NewLocationName(projectId, locationId string) NameDetail {
	return newNameDetail(ResourceTypeLocation, nameDefnProject.Value(projectId), nameDefnLocation.Value(locationId))
}

func NewInstanceName(projectId, locationId, instanceId string) NameDetail {
	return newNameDetail(ResourceTypeInstance, nameDefnProject.Value(projectId), nameDefnLocation.Value(locationId), nameDefnInstance.Value(instanceId))
}

func NewBackupName(projectId, locationId, backupId string) NameDetail {
	return newNameDetail(ResourceTypeBackup, nameDefnProject.Value(projectId), nameDefnLocation.Value(locationId), nameDefnBackup.Value(backupId))
}

func NewClusterName(projectId, locationId, clusterId string) NameDetail {
	return newNameDetail(ResourceTypeCluster, nameDefnProject.Value(projectId), nameDefnLocation.Value(locationId), nameDefnCluster.Value(clusterId))
}

func NewGlobalAddressName(projectId, addressId string) NameDetail {
	return newNameDetail(ResourceTypeGlobalAddress, nameDefnProject.Value(projectId), nameDefnAddress.Value(addressId))
}

func NewServiceName(projectId, serviceId string) NameDetail {
	return newNameDetail(ResourceTypeService, nameDefnProject.Value(projectId), nameDefnService.Value(serviceId))
}

func NewRegionName(projectId, regionId string) NameDetail {
	return newNameDetail(ResourceTypeRegion, nameDefnProject.Value(projectId), nameDefnRegion.Value(regionId))
}

func NewRouterName(projectId, regionId, routerId string) NameDetail {
	return newNameDetail(ResourceTypeRouter, nameDefnProject.Value(projectId), nameDefnRegion.Value(regionId), nameDefnRouter.Value(routerId))
}

func NewRegionalAddressName(projectId, regionId, addressId string) NameDetail {
	return newNameDetail(ResourceTypeRegionalAddress, nameDefnProject.Value(projectId), nameDefnRegion.Value(regionId), nameDefnAddresses.Value(addressId))
}

func NewRegionalOperationName(projectId, regionId, operationId string) NameDetail {
	return newNameDetail(ResourceTypeGlobalNetwork, nameDefnProject.Value(projectId), nameDefnRegion.Value(regionId), nameDefnOperation.Value(operationId))
}

func NewSubnetworkName(projectId, regionId, subnetworkId string) NameDetail {
	return newNameDetail(ResourceTypeSubnetwork, nameDefnProject.Value(projectId), nameDefnRegion.Value(regionId), nameDefnSubnetwork.Value(subnetworkId))
}

// parse ---------------------------------------------------------------

func MustParseNameDetail(path string) NameDetail {
	return util.Must(ParseNameDetail(path))
}

func ParseNameDetail(path string) (NameDetail, error) {
	if strings.Contains(path, "://") {
		if i := strings.Index(path, "/projects/"); i > -1 {
			path = path[i+1:]
		}
	}
	pathParts := strings.Split(path, "/")
	var parts []namePart
	for {
		var match *namePartDefn
		for _, defn := range allNamePartDefns {
			p := defn.Match(pathParts)
			if p != nil {
				match = &defn
				parts = append(parts, *p)
				pathParts = pathParts[len(defn.parts):]
				break
			}
		}
		if match == nil {
			return NameDetail{}, NewErrInvalidNameSyntax(path)
		}
		if len(pathParts) == 0 {
			break
		}
	}
	var matchesValidSequence *ResourceType
	for rt, validSequence := range validSequences {
		if len(validSequence) != len(parts) {
			continue
		}
		matches := true
		for i := range validSequence {
			if validSequence[i].format != parts[i].defn.format {
				matches = false
				break
			}
		}
		if matches {
			matchesValidSequence = &rt
			break
		}
	}
	if matchesValidSequence == nil {
		return NameDetail{}, fmt.Errorf("gcp name %s has valid syntax but has no valid sequence", path)
	}
	return NameDetail{
		parts:        parts,
		resourceType: *matchesValidSequence,
	}, nil
}

func newNameDetail(resourceType ResourceType, parts ...namePart) NameDetail {
	return NameDetail{
		parts:        parts,
		resourceType: resourceType,
	}
}

func (n NameDetail) Equal(other NameDetail) bool {
	if n.resourceType != other.resourceType {
		return false
	}
	if len(n.parts) != len(other.parts) {
		return false
	}
	for i := range n.parts {
		if n.parts[i].defn.format != other.parts[i].defn.format || n.parts[i].value != other.parts[i].value {
			return false
		}
	}
	return true
}

// ProjectId returns the project id of the name. It is the value of the part with format projects/%s.
func (n NameDetail) ProjectId() string {
	for _, p := range n.parts {
		if p.defn.format == nameDefnProject.format {
			return p.value
		}
	}
	return ""
}

// LocationRegionId returns the location or region id of the name. It is the value of the part with format locations/%s or regions/%s.
// For names w/out location/region part, it returns empty string.
func (n NameDetail) LocationRegionId() string {
	for _, p := range n.parts {
		if p.defn.format == nameDefnLocation.format || p.defn.format == nameDefnRegion.format {
			return p.value
		}
	}
	return ""
}

// ResourceId returns the id of the resource, ie network id, location id, instance id, etc. It is the value of the last part of the name.
func (n NameDetail) ResourceId() string {
	return n.parts[len(n.parts)-1].value
}

// projects/%s/global/networks/%s
// projects/%s/locations/%s
// projects/%s/locations/%s/instances/%s
// projects/%s/locations/%s/backups/%s
// projects/%s/locations/%s/clusters/%s
// projects/%s/address/%s
// projects/%s/services/%s
// projects/%s/regions/%s
// projects/%s/regions/%s/routers/%s
// projects/%s/regions/%s/addresses/%s
// projects/%s/regions/%s/subnetworks/%s

type ResourceType string

const (
	ResourceTypeGlobalNetwork     ResourceType = "globalNetwork"
	ResourceTypeGlobalOperation   ResourceType = "globalOperation"
	ResourceTypeRegionalOperation ResourceType = "regionalOperation"
	ResourceTypeLocation          ResourceType = "location"
	ResourceTypeInstance          ResourceType = "instance"
	ResourceTypeBackup            ResourceType = "backup"
	ResourceTypeCluster           ResourceType = "cluster"
	ResourceTypeGlobalAddress     ResourceType = "globalAddress"
	ResourceTypeService           ResourceType = "service"
	ResourceTypeRegion            ResourceType = "region"
	ResourceTypeRouter            ResourceType = "router"
	ResourceTypeRegionalAddress   ResourceType = "regionalAddress"
	ResourceTypeSubnetwork        ResourceType = "subnetwork"
)

func (n NameDetail) ResourceType() ResourceType {
	return n.resourceType
}

func (n NameDetail) PrefixWith(prefix string) string {
	prefix = strings.TrimSuffix(prefix, "/")
	return fmt.Sprintf("%s/%s", prefix, n.String())
}

const prefixGoogleApisComputeV1 = "https://www.googleapis.com/compute/v1/"

func (n NameDetail) PrefixWithGoogleApisComputeV1() string {
	return n.PrefixWith(prefixGoogleApisComputeV1)
}

func (n NameDetail) String() string {
	var sb strings.Builder
	for i, p := range n.parts {
		if i > 0 {
			sb.WriteString("/")
		}
		sb.WriteString(p.String())
	}
	return sb.String()
}

// name part defn =========================================================

type namePartDefn struct {
	format string
	parts  []string
}

func newNamePartDefn(format string) namePartDefn {
	parts := strings.Split(format, "/")
	return namePartDefn{
		format: format,
		parts:  parts,
	}
}

func (n namePartDefn) Match(parts []string) *namePart {
	if len(parts) < len(n.parts) {
		return nil
	}
	var value string
	for i := range n.parts {
		if n.parts[i] == "%s" {
			value = parts[i]
		} else if n.parts[i] != parts[i] {
			return nil
		}
	}
	return &namePart{
		defn:  &n,
		value: value,
	}
}

func (n namePartDefn) Value(val string) namePart {
	return namePart{
		defn:  &n,
		value: val,
	}
}

var (
	nameDefnProject       = newNamePartDefn("projects/%s")
	nameDefnGlobalNetwork = newNamePartDefn("global/networks/%s")
	nameDefnOperation     = newNamePartDefn("operations/%s")
	nameDefnLocation      = newNamePartDefn("locations/%s")
	nameDefnInstance      = newNamePartDefn("instances/%s")
	nameDefnBackup        = newNamePartDefn("backups/%s")
	nameDefnCluster       = newNamePartDefn("clusters/%s")
	nameDefnAddress       = newNamePartDefn("address/%s")
	nameDefnAddresses     = newNamePartDefn("addresses/%s")
	nameDefnService       = newNamePartDefn("services/%s")
	nameDefnRegion        = newNamePartDefn("regions/%s")
	nameDefnRouter        = newNamePartDefn("routers/%s")
	nameDefnSubnetwork    = newNamePartDefn("subnetworks/%s")
)

var allNamePartDefns = []namePartDefn{
	nameDefnProject,
	nameDefnGlobalNetwork,
	nameDefnOperation,
	nameDefnLocation,
	nameDefnInstance,
	nameDefnBackup,
	nameDefnCluster,
	nameDefnAddress,
	nameDefnAddresses,
	nameDefnService,
	nameDefnRegion,
	nameDefnRouter,
	nameDefnSubnetwork,
}

// projects/%s/global/networks/%s
// projects/%s/locations/%s
// projects/%s/locations/%s/instances/%s
// projects/%s/locations/%s/backups/%s
// projects/%s/locations/%s/clusters/%s
// projects/%s/address/%s
// projects/%s/services/%s
// projects/%s/regions/%s
// projects/%s/regions/%s/routers/%s
// projects/%s/regions/%s/addresses/%s
// projects/%s/regions/%s/subnetworks/%s
// projects/%s/operations/%s

var validSequences = map[ResourceType][]namePartDefn{
	ResourceTypeGlobalNetwork:     {nameDefnProject, nameDefnGlobalNetwork},
	ResourceTypeGlobalOperation:   {nameDefnProject, nameDefnOperation},
	ResourceTypeRegionalOperation: {nameDefnProject, nameDefnRegion, nameDefnOperation},
	ResourceTypeLocation:          {nameDefnProject, nameDefnLocation},
	ResourceTypeInstance:          {nameDefnProject, nameDefnLocation, nameDefnInstance},
	ResourceTypeBackup:            {nameDefnProject, nameDefnLocation, nameDefnBackup},
	ResourceTypeCluster:           {nameDefnProject, nameDefnLocation, nameDefnCluster},
	ResourceTypeGlobalAddress:     {nameDefnProject, nameDefnAddress},
	ResourceTypeService:           {nameDefnProject, nameDefnService},
	ResourceTypeRegion:            {nameDefnProject, nameDefnRegion},
	ResourceTypeRouter:            {nameDefnProject, nameDefnRegion, nameDefnRouter},
	ResourceTypeRegionalAddress:   {nameDefnProject, nameDefnRegion, nameDefnAddresses},
	ResourceTypeSubnetwork:        {nameDefnProject, nameDefnRegion, nameDefnSubnetwork},
}

// name part =========================================================

type namePart struct {
	defn  *namePartDefn
	value string
}

func (p namePart) String() string {
	return fmt.Sprintf(p.defn.format, p.value)
}

// prefixes:
// https://www.googleapis.com/compute/v1/
