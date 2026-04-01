package mock2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/elliotchance/pie/v2"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/servicenetworking/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
createTime: '2026-02-23T09:51:00.946032849Z'
description: dabc1234-c123-123f-b2fa-1d12342b2e80
fileShares:
- capacityGb: '1024'
  name: vol1
name: projects/my-project/locations/us-central1-a/instances/my-instance
networks:
- connectMode: PRIVATE_SERVICE_ACCESS
  ipAddresses:
  - 10.250.4.2
  modes:
  - MODE_IPV4
  network: projects/my-project/global/networks/my-network
  reservedIpRange: 10.250.4.0/29
performanceLimits:
  maxIops: '600'
  maxReadIops: '600'
  maxReadThroughputBps: '104857600'
  maxWriteIops: '1000'
  maxWriteThroughputBps: '104857600'
satisfiesPzi: true
satisfiesPzs: false
state: READY
tier: BASIC_HDD
*/

// private methods ============================================================================

func (s *store) getFilestoreNoLock(name string) (*filestorepb.Instance, error) {
	for _, item := range s.filestores.items {
		if item.Name.EqualString(name) {
			return item.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("filestore %s not found", name)
}

// Filestore =====================================================================================

func (s *store) GetFilestoreInstance(ctx context.Context, req *filestorepb.GetInstanceRequest, _ ...gax.CallOption) (*filestorepb.Instance, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	fs, err := s.getFilestoreNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(fs)
}

func (s *store) CreateFilestoreInstance(ctx context.Context, req *filestorepb.CreateInstanceRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*filestorepb.Instance], error) {
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
	fsName := gcputil.NewInstanceName(parentName.ProjectId(), parentName.LocationRegionId(), req.InstanceId)
	_, err = s.getFilestoreNoLock(fsName.String())
	if err == nil {
		return nil, gcpmeta.NewBadRequestError("filestore %s already exists", fsName.String())
	}

	// network validation
	if len(req.Instance.Networks) != 1 {
		return nil, gcpmeta.NewBadRequestError("only one network specification is required")
	}
	netName, err := gcputil.ParseNameDetail(req.Instance.Networks[0].Network)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid network specification: %v", err)
	}
	_, err = s.GetNetworkNoLock(netName.ProjectId(), netName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found: %v", netName.String(), err)
	}

	// network PSA config and address range validation
	if req.Instance.Networks[0].ConnectMode != filestorepb.NetworkConfig_PRIVATE_SERVICE_ACCESS {
		return nil, gcpmeta.NewBadRequestError("in mock only PRIVATE_SERVICE_ACCESS mode is supported")
	}
	if req.Instance.Networks[0].ReservedIpRange == "" {
		return nil, gcpmeta.NewBadRequestError("reserved ip range is required")
	}

	addrName, err := gcputil.ParseNameDetail(req.Instance.Networks[0].ReservedIpRange)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid reserved ip range: %v", err)
	}
	addr, err := s.getAddressNoLock(addrName.ProjectId(), addrName.LocationRegionId(), addrName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("reserved ip range %s not found: %v", addrName.String(), err)
	}
	if !netName.EqualString(addr.GetNetwork()) {
		return nil, gcpmeta.NewBadRequestError("reserved ip range %q belongs to network %q, while filestore is in network %q", addrName.String(), addr.GetNetwork(), netName.String())
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

	// validate file shares

	if len(req.Instance.FileShares) != 1 {
		return nil, gcpmeta.NewBadRequestError("only one file share specification is required")
	}

	// create filestore

	fs, err := util.Clone(req.Instance)
	if err != nil {
		return nil, fmt.Errorf("%w failed to clone filestore: %w", common.ErrLogical, err)
	}

	fs.Name = fsName.String()

	// add a filestore reactor to handle filestore state and operation change
	fs.State = filestorepb.Instance_CREATING

	// This is simplified, first ip of the range is assigned, good enough for the tests
	// Ideally, a separate address space for this range should be maintained, and instead an ip allocated from it
	// either by having addressSpace instance, which seems to be over-do, or by iterating over all filestore instances
	// and finding an ip from the range that's not used, which seems to be to many computations
	fs.Networks[0].IpAddresses = []string{addr.GetAddress()}

	s.filestores.Add(fs, fsName)

	// Do not put the result right away in the op, since it's a marshaled copy into pb.Any, so further mutations
	// on the filestore instance will not get reflected to the operation result. Instead, when operation
	// is mutated to done with reactor, also mutate the filestore instance in the mock and only then put it to the op
	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), fsName)
	if err := b.WithCommonMetadata(fsName, "create"); err != nil {
		return nil, fmt.Errorf("%w: failed setting create filestore operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*filestorepb.Instance](b.GetOperationPB()), nil
}

func (s *store) UpdateFilestoreInstance(ctx context.Context, req *filestorepb.UpdateInstanceRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*filestorepb.Instance], error) {
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

	fsName, err := gcputil.ParseNameDetail(req.Instance.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid filestore instance name: %v", err)
	}

	fs, err := s.getFilestoreNoLock(req.Instance.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("filestore %s not found: %v", req.Instance.Name, err)
	}

	err = UpdateMask(fs, req.Instance, req.UpdateMask)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("filestore %s update failed: %v", req.Instance.Name, err)
	}

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), fsName)
	if err := b.WithCommonMetadata(fsName, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting update filestore operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*filestorepb.Instance](b.GetOperationPB()), nil
}

func (s *store) DeleteFilestoreInstance(ctx context.Context, req *filestorepb.DeleteInstanceRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("file store instance name is required")
	}
	fsName, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store instance name: %v", err)
	}

	fs, err := s.getFilestoreNoLock(req.Name)
	if err != nil {
		return nil, err
	}

	fs.State = filestorepb.Instance_DELETING

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), fsName)
	if err := b.WithCommonMetadata(fsName, "delete"); err != nil {
		return nil, fmt.Errorf("%w: failed setting delete filestore operation metadata: %w", common.ErrLogical, err)
	}

	s.longRunningOperations.Add(b, opName)

	return b.BuildVoidOperation(), nil
}

// Backup =====================================================================================

func (s *store) getFilestoreBackupNoLock(backupName string) (*filestorepb.Backup, error) {
	nd, err := gcputil.ParseNameDetail(backupName)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup name: %v", err)
	}
	b, found := s.backups.FindByName(nd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("file store backup %s not found", nd.String())
	}
	return b, nil
}

func (s *store) UpdateFilestoreBackup(ctx context.Context, req *filestorepb.UpdateBackupRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*filestorepb.Backup], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Backup == nil {
		return nil, gcpmeta.NewBadRequestError("backup is required")
	}
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, gcpmeta.NewBadRequestError("update mask is required")
	}

	backupNd, err := gcputil.ParseNameDetail(req.Backup.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup backup name: %v", err)
	}
	backup, found := s.backups.FindByName(backupNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("backup %s not found", req.Backup.Name)
	}

	err = UpdateMask(backup, req.Backup, req.UpdateMask)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("filestore %s update failed: %v", req.Backup.Name, err)
	}

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), backupNd)
	if err := b.WithCommonMetadata(backupNd, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting update filestore backup operation metadata: %w", common.ErrLogical, err)
	}
	b.WithDone(true)

	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*filestorepb.Backup](b.GetOperationPB()), nil
}

func (s *store) GetFilestoreBackup(ctx context.Context, req *filestorepb.GetBackupRequest, _ ...gax.CallOption) (*filestorepb.Backup, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	b, err := s.getFilestoreBackupNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(b)
}

func (s *store) ListFilestoreBackups(ctx context.Context, req *filestorepb.ListBackupsRequest, _ ...gax.CallOption) gcpclient.Iterator[*filestorepb.Backup] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*filestorepb.Backup]{
			err: ctx.Err(),
		}
	}

	var err error

	list := s.backups
	if req.Parent != "" {
		parentNd, err := gcputil.ParseNameDetail(req.Parent)
		if err != nil {
			return &iteratorMocked[*filestorepb.Backup]{
				err: gcpmeta.NewBadRequestError("invalid file store backup parent name: %v", err),
			}
		}
		list = list.FilterByCallback(func(item FilterableListItem[*filestorepb.Backup]) bool {
			if item.Name.ProjectId() != parentNd.ProjectId() {
				return false
			}
			// "-" is the GCP wildcard meaning "all locations"
			if parentNd.LocationRegionId() == "-" {
				return true
			}
			return item.Name.LocationRegionId() == parentNd.LocationRegionId()
		})
	}
	if req.Filter != "" {
		list, err = list.FilterByExpression(&req.Filter)
		if err != nil {
			return &iteratorMocked[*filestorepb.Backup]{
				err: gcpmeta.NewBadRequestError("invalid file store backup filter: %v", err),
			}
		}
	}

	return list.ToIterator()
}

func (s *store) CreateFilestoreBackup(ctx context.Context, req *filestorepb.CreateBackupRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*filestorepb.Backup], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	parentNd, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup parent name: %v", err)
	}
	if parentNd.ResourceType() != gcputil.ResourceTypeLocation && parentNd.ResourceType() != gcputil.ResourceTypeRegion {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup parent name type %s, expected location or region", parentNd.ResourceType())
	}

	if req.BackupId == "" {
		return nil, gcpmeta.NewBadRequestError("backup id is required")
	}

	backupNd := gcputil.NewBackupName(parentNd.ProjectId(), parentNd.LocationRegionId(), req.BackupId)

	_, found := s.backups.FindByName(backupNd)
	if found {
		return nil, gcpmeta.NewBadRequestError("backup %s already exists", backupNd.String())
	}

	// validate source filestore instance & its file share

	if req.Backup.SourceInstance == "" {
		return nil, gcpmeta.NewBadRequestError("backup source instance is required")
	}
	fs, err := s.getFilestoreNoLock(req.Backup.SourceInstance)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup source instance: %v", err)
	}
	if len(fs.FileShares) != 1 {
		return nil, gcpmeta.NewBadRequestError("%v: expected 1 file share but found %d", common.ErrLogical, len(fs.FileShares))
	}
	if fs.FileShares[0].Name != req.Backup.SourceFileShare {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup source file share %s", req.Backup.SourceFileShare)
	}

	backup, err := util.Clone(req.Backup)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("%v: failed to clone backup: %v", common.ErrLogical, err)
	}

	backup.Name = backupNd.String()
	backup.State = filestorepb.Backup_CREATING
	backup.CreateTime = timestamppb.Now()
	backup.CapacityGb = fs.FileShares[0].CapacityGb
	backup.StorageBytes = backup.CapacityGb * 1024 * 1024 / 10
	backup.SourceInstanceTier = fs.Tier
	backup.DownloadBytes = backup.StorageBytes
	backup.FileSystemProtocol = fs.Protocol

	s.backups.Add(backup, backupNd)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), backupNd)
	if err := b.WithCommonMetadata(backupNd, "create"); err != nil {
		return nil, fmt.Errorf("%w: failed setting create filestore backup operation metadata: %w", common.ErrLogical, err)
	}

	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*filestorepb.Backup](b.GetOperationPB()), nil
}

func (s *store) DeleteFilestoreBackup(ctx context.Context, req *filestorepb.DeleteBackupRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	backupNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid file store backup backup name: %v", err)
	}
	backup, found := s.backups.FindByName(backupNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("backup %s not found", req.Name)
	}

	backup.State = filestorepb.Backup_DELETING

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), backupNd)
	if err := b.WithCommonMetadata(backupNd, "delete"); err != nil {
		return nil, fmt.Errorf("%w: failed setting delete filestore backup operation metadata: %w", common.ErrLogical, err)
	}

	s.longRunningOperations.Add(b, opName)

	return b.BuildVoidOperation(), nil
}
