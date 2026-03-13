package mock2

import (
	"context"
	"fmt"
	"slices"

	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/elliotchance/pie/v2"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/servicenetworking/v1"
)

/*
authEnabled: true
authorizedNetwork: projects/my-project/global/networks/my-network
availableMaintenanceVersions:
- '20251110_01_00'
connectMode: PRIVATE_SERVICE_ACCESS
createTime: '2025-03-31T17:44:56.236382111Z'
currentLocationId: europe-west1-c
host: 172.16.0.1
locationId: europe-west1-c
maintenanceVersion: '20251007_00_00'
memorySizeGb: 2
name: projects/my-project/locations/europe-west1/instances/my-instance
nodes:
- id: node-0
  zone: europe-west1-c
persistenceConfig:
  persistenceMode: DISABLED
persistenceIamIdentity: serviceAccount:service-12341234234@cloud-redis.iam.gserviceaccount.com
port: 6378
readReplicasMode: READ_REPLICAS_DISABLED
redisConfigs:
  maxmemory-policy: noeviction
redisVersion: REDIS_7_0
reservedIpRange: 172.16.0.0/29
satisfiesPzi: true
serverCaCerts:
- cert: |-
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  createTime: '2025-03-31T17:45:04.218873Z'
  expireTime: '2035-03-29T17:45:04.073Z'
  serialNumber: '0'
  sha1Fingerprint: 123a3b1234b4b02e3a21a12e2a2c8c123f21d7a12
state: READY
tier: BASIC
transitEncryptionMode: SERVER_AUTHENTICATION
*/

// private methods ============================================================================

func (s *store) getRedisInstanceNoLock(name string) (*redispb.Instance, error) {
	for _, item := range s.redisInstances.items {
		if item.Name.EqualString(name) {
			return item.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("redisInstance %s not found", name)
}

// RedisInstance public client methods ========================================================

func (s *store) CreateRedisInstance(ctx context.Context, req *redispb.CreateInstanceRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*redispb.Instance], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	// parent validation
	parentName, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid parent: %v", err)
	}
	if parentName.ResourceType() != gcputil.ResourceTypeLocation {
		return nil, gcpmeta.NewBadRequestError("invalid resource type, expected location got %q", parentName.ResourceType())
	}

	// name validation
	if req.InstanceId == "" {
		return nil, gcpmeta.NewBadRequestError("instanceId is required")
	}
	riName := gcputil.NewInstanceName(parentName.ProjectId(), parentName.LocationRegionId(), req.InstanceId)
	_, err = s.getRedisInstanceNoLock(riName.String())
	if err == nil {
		return nil, gcpmeta.NewBadRequestError("redisInstance %s already exists", riName.String())
	}

	// redis values validation
	validRedisVersions := []string{"REDIS_3_2", "REDIS_4_0", "REDIS_5_0", "REDIS_6_X"}
	if !slices.Contains(validRedisVersions, req.Instance.RedisVersion) {
		return nil, gcpmeta.NewBadRequestError("invalid redisVersion %q, must be one of %v", req.Instance.RedisVersion, validRedisVersions)
	}

	// network validation
	if req.Instance.ConnectMode != redispb.Instance_PRIVATE_SERVICE_ACCESS {
		return nil, gcpmeta.NewBadRequestError("redisInstance connect mode PSA is only supported in mock")
	}

	if req.Instance.AuthorizedNetwork == "" {
		return nil, gcpmeta.NewBadRequestError("redisInstance authorizedNetwork is required")
	}
	netName, err := gcputil.ParseNameDetail(req.Instance.AuthorizedNetwork)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid authorized network name: %v", err)
	}
	_, err = s.getNetworkNoLock(netName.ProjectId(), netName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewNotFoundError("authorized network %s not found", netName.String())
	}

	// PSA address range validation
	if req.Instance.ReservedIpRange == "" {
		return nil, gcpmeta.NewBadRequestError("reservedIpRange is required")
	}
	// not sure about this, so supporting both
	// first try if ReservedIpRange is valid address self link
	addrName, err := gcputil.ParseNameDetail(req.Instance.ReservedIpRange)
	if err != nil {
		// if ReservedIpRange is not a valid address self link, construct the addr link assuming it's just the address id
		addrName = gcputil.NewRegionalAddressName(riName.ProjectId(), riName.LocationRegionId(), req.Instance.ReservedIpRange)
	}
	addr, err := s.getAddressNoLock(addrName.ProjectId(), addrName.LocationRegionId(), addrName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewNotFoundError("address range %s not found", addrName.String())
	}
	if !netName.EqualString(addr.GetNetwork()) {
		return nil, gcpmeta.NewBadRequestError("reserved ip range %q belongs to network %q, while redisInstance is in network %q", addrName.String(), addr.GetNetwork(), netName.String())
	}
	if addr.GetPurpose() != "VPC_PEERING" {
		return nil, gcpmeta.NewBadRequestError("only address range of purpose VPC_PEERING can be used, but got %q", addr.GetPurpose())
	}

	serviceConnections := s.serviceConnections.
		FilterByCallback(func(item FilterableListItem[*servicenetworking.Connection]) bool {
			return netName.EqualString(item.Obj.Network)
		}).
		FilterByCallback(func(item FilterableListItem[*servicenetworking.Connection]) bool {
			return pie.Contains(item.Obj.ReservedPeeringRanges, addr.GetSelfLink())
		})
	if serviceConnections.Len() == 0 {
		return nil, gcpmeta.NewNotFoundError("service connection in network %s linked to address %s not found", netName.String(), addr.GetSelfLink())
	}

	// create redis instance
	ri, err := util.Clone(req.Instance)
	if err != nil {
		return nil, fmt.Errorf("%w failed to clone redis instance: %w", common.ErrLogical, err)
	}

	ri.Name = riName.String()
	ri.State = redispb.Instance_CREATING
	// simplified takes first ip of the range
	ri.Host = addr.GetAddress()
	ri.Port = 6379

	s.redisInstances.Add(ri, riName)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), riName)
	if err := b.WithRedisInstanceMetadata(riName, "create"); err != nil {
		return nil, fmt.Errorf("%w: failed setting create redisInstance operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*redispb.Instance](b.GetOperationPB()), nil
}

func (s *store) GetRedisInstance(ctx context.Context, req *redispb.GetInstanceRequest, _ ...gax.CallOption) (*redispb.Instance, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	ri, err := s.getRedisInstanceNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(ri)
}

func (s *store) UpdateRedisInstance(ctx context.Context, req *redispb.UpdateInstanceRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*redispb.Instance], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Instance == nil {
		return nil, gcpmeta.NewBadRequestError("instance is required")
	}
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, gcpmeta.NewBadRequestError("update mask is required")
	}

	riName, err := gcputil.ParseNameDetail(req.Instance.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid redisInstance instance name: %v", err)
	}

	ri, err := s.getRedisInstanceNoLock(req.Instance.Name)
	if err != nil {
		return nil, gcpmeta.NewNotFoundError("redisInstance %s not found", req.Instance.Name)
	}

	err = UpdateMask(ri, req.Instance, req.UpdateMask)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("redisInstance %s update failed: %v", riName.String(), err)
	}

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), riName)
	if err := b.WithRedisInstanceMetadata(riName, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting update redisInstance operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*redispb.Instance](b.GetOperationPB()), nil
}

func (s *store) UpgradeRedisInstance(ctx context.Context, req *redispb.UpgradeInstanceRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*redispb.Instance], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	riName, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid redisInstance instance name: %v", err)
	}

	ri, err := s.getRedisInstanceNoLock(riName.String())
	if err != nil {
		return nil, gcpmeta.NewNotFoundError("redisInstance %s not found", riName.String())
	}

	ri.RedisVersion = req.RedisVersion
	ri.State = redispb.Instance_MAINTENANCE

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), riName)
	if err := b.WithRedisInstanceMetadata(riName, "upgrade"); err != nil {
		return nil, fmt.Errorf("%w: failed setting upgrade redisInstance operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*redispb.Instance](b.GetOperationPB()), nil
}

func (s *store) DeleteRedisInstance(ctx context.Context, req *redispb.DeleteInstanceRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	riName, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid redisInstance instance name: %v", err)
	}

	ri, err := s.getRedisInstanceNoLock(riName.String())
	if err != nil {
		return nil, gcpmeta.NewNotFoundError("redisInstance %s not found", riName.String())
	}

	ri.State = redispb.Instance_DELETING

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), riName)
	if err := b.WithRedisInstanceMetadata(riName, "delete"); err != nil {
		return nil, fmt.Errorf("%w: failed setting delete redisInstance operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return b.BuildVoidOperation(), nil
}
