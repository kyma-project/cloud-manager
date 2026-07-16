package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
)

// NasFileSystemEntry is the stored representation of a NAS file system.
type NasFileSystemEntry struct {
	FileSystemId   string
	FileSystemType string
	ProtocolType   string
	StorageType    string
	Status         string
	ZoneId         string
	MeteredSize    int64
	Capacity       int64
}

// NasMountTargetEntry is the stored representation of a NAS mount target.
type NasMountTargetEntry struct {
	MountTargetDomain string
	FileSystemId      string
	NetworkType       string
	VpcId             string
	VSwitchId         string
	AccessGroup       string
	Status            string
}

// NasAccessGroupEntry is the stored representation of a NAS permission group.
type NasAccessGroupEntry struct {
	AccessGroupName string
	AccessGroupType string
	FileSystemType  string
	Rules           []string
}

type nasStore struct {
	m sync.Mutex

	fileSystems  []*NasFileSystemEntry
	mountTargets []*NasMountTargetEntry
	accessGroups []*NasAccessGroupEntry

	fileSystemErrors map[string]error
}

func newNasStore() *nasStore {
	return &nasStore{
		fileSystemErrors: map[string]error{},
	}
}

// === Config side (test seeding) =============================================

func (s *nasStore) AddNasFileSystem(id, protocolType, storageType, zoneId string) *NasFileSystemEntry {
	s.m.Lock()
	defer s.m.Unlock()
	if id == "" {
		id = uuid.NewString()[:8]
	}
	entry := &NasFileSystemEntry{
		FileSystemId:   id,
		FileSystemType: "standard",
		ProtocolType:   protocolType,
		StorageType:    storageType,
		ZoneId:         zoneId,
		Status:         "Running",
	}
	s.fileSystems = append(s.fileSystems, entry)
	return entry
}

func (s *nasStore) SetNasFileSystemError(fileSystemId string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.fileSystemErrors, fileSystemId)
	} else {
		s.fileSystemErrors[fileSystemId] = err
	}
}

// === nfsinstance.Client: file system ========================================

func (s *nasStore) DescribeFileSystem(ctx context.Context, fileSystemId string) (*alicloudnfsinstanceclient.FileSystemInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.fileSystemErrors[fileSystemId]; ok {
		return nil, err
	}
	idx := pie.FindFirstUsing(s.fileSystems, func(f *NasFileSystemEntry) bool { return f.FileSystemId == fileSystemId })
	if idx == -1 {
		return nil, nil
	}
	f := s.fileSystems[idx]
	return &alicloudnfsinstanceclient.FileSystemInfo{
		FileSystemId:   f.FileSystemId,
		FileSystemType: f.FileSystemType,
		ProtocolType:   f.ProtocolType,
		StorageType:    f.StorageType,
		Status:         f.Status,
		ZoneId:         f.ZoneId,
		MeteredSize:    f.MeteredSize,
		Capacity:       f.Capacity,
	}, nil
}

func (s *nasStore) CreateFileSystem(ctx context.Context, protocolType, storageType, zoneId string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	entry := s.AddNasFileSystem("", protocolType, storageType, zoneId)
	return entry.FileSystemId, nil
}

func (s *nasStore) DeleteFileSystem(ctx context.Context, fileSystemId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.fileSystemErrors[fileSystemId]; ok {
		return err
	}
	idx := pie.FindFirstUsing(s.fileSystems, func(f *NasFileSystemEntry) bool { return f.FileSystemId == fileSystemId })
	if idx == -1 {
		return fmt.Errorf("nas file system %s not found", fileSystemId)
	}
	for _, mt := range s.mountTargets {
		if mt.FileSystemId == fileSystemId {
			return fmt.Errorf("nas file system %s has dependent mount targets", fileSystemId)
		}
	}
	s.fileSystems = append(s.fileSystems[:idx], s.fileSystems[idx+1:]...)
	return nil
}

// === nfsinstance.Client: mount targets ======================================

func (s *nasStore) DescribeMountTargets(ctx context.Context, fileSystemId string) ([]alicloudnfsinstanceclient.MountTargetInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	var out []alicloudnfsinstanceclient.MountTargetInfo
	for _, mt := range s.mountTargets {
		if mt.FileSystemId == fileSystemId {
			out = append(out, alicloudnfsinstanceclient.MountTargetInfo{
				MountTargetDomain: mt.MountTargetDomain,
				NetworkType:       mt.NetworkType,
				VpcId:             mt.VpcId,
				VSwitchId:         mt.VSwitchId,
				AccessGroup:       mt.AccessGroup,
				Status:            mt.Status,
			})
		}
	}
	return out, nil
}

func (s *nasStore) CreateMountTarget(ctx context.Context, fileSystemId, vpcId, vSwitchId, accessGroupName string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if idx := pie.FindFirstUsing(s.fileSystems, func(f *NasFileSystemEntry) bool { return f.FileSystemId == fileSystemId }); idx == -1 {
		return "", fmt.Errorf("nas file system %s not found", fileSystemId)
	}
	domain := fmt.Sprintf("%s-%s.nas.aliyuncs.com", fileSystemId, uuid.NewString()[:4])
	s.mountTargets = append(s.mountTargets, &NasMountTargetEntry{
		MountTargetDomain: domain,
		FileSystemId:      fileSystemId,
		NetworkType:       "Vpc",
		VpcId:             vpcId,
		VSwitchId:         vSwitchId,
		AccessGroup:       accessGroupName,
		Status:            "Active",
	})
	return domain, nil
}

func (s *nasStore) DeleteMountTarget(ctx context.Context, fileSystemId, mountTargetDomain string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.mountTargets, func(mt *NasMountTargetEntry) bool {
		return mt.FileSystemId == fileSystemId && mt.MountTargetDomain == mountTargetDomain
	})
	if idx == -1 {
		return fmt.Errorf("nas mount target %s not found", mountTargetDomain)
	}
	s.mountTargets = append(s.mountTargets[:idx], s.mountTargets[idx+1:]...)
	return nil
}

// === nfsinstance.Client: access groups ======================================

func (s *nasStore) DescribeAccessGroups(ctx context.Context, accessGroupName string) ([]alicloudnfsinstanceclient.AccessGroupInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	var out []alicloudnfsinstanceclient.AccessGroupInfo
	for _, ag := range s.accessGroups {
		if accessGroupName == "" || ag.AccessGroupName == accessGroupName {
			out = append(out, alicloudnfsinstanceclient.AccessGroupInfo{
				AccessGroupName: ag.AccessGroupName,
				AccessGroupType: ag.AccessGroupType,
				FileSystemType:  ag.FileSystemType,
			})
		}
	}
	return out, nil
}

func (s *nasStore) CreateAccessGroup(ctx context.Context, accessGroupName, description string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if pie.FindFirstUsing(s.accessGroups, func(ag *NasAccessGroupEntry) bool { return ag.AccessGroupName == accessGroupName }) != -1 {
		return nil
	}
	s.accessGroups = append(s.accessGroups, &NasAccessGroupEntry{
		AccessGroupName: accessGroupName,
		AccessGroupType: "Vpc",
		FileSystemType:  "standard",
	})
	return nil
}

func (s *nasStore) DeleteAccessGroup(ctx context.Context, accessGroupName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.accessGroups, func(ag *NasAccessGroupEntry) bool { return ag.AccessGroupName == accessGroupName })
	if idx == -1 {
		return nil
	}
	s.accessGroups = append(s.accessGroups[:idx], s.accessGroups[idx+1:]...)
	return nil
}

func (s *nasStore) DescribeAccessRules(ctx context.Context, accessGroupName string) ([]string, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.accessGroups, func(ag *NasAccessGroupEntry) bool { return ag.AccessGroupName == accessGroupName })
	if idx == -1 {
		return nil, nil
	}
	out := make([]string, len(s.accessGroups[idx].Rules))
	copy(out, s.accessGroups[idx].Rules)
	return out, nil
}

func (s *nasStore) CreateAccessRule(ctx context.Context, accessGroupName, sourceCidrIp string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.accessGroups, func(ag *NasAccessGroupEntry) bool { return ag.AccessGroupName == accessGroupName })
	if idx == -1 {
		return fmt.Errorf("nas access group %s not found", accessGroupName)
	}
	s.accessGroups[idx].Rules = append(s.accessGroups[idx].Rules, sourceCidrIp)
	return nil
}
