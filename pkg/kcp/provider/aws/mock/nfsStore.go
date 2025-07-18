package mock

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sync"
)

type NfsConfig interface {
	SetFileSystemLifeCycleState(id string, state efstypes.LifeCycleState)
	GetFileSystemById(id string) *efstypes.FileSystemDescription
	SetFileSystemStatusCapacityById(id string, capacity resource.Quantity)
}

type mountTargetItem struct {
	desc efstypes.MountTargetDescription
	sg   []string
}

type nfsStore struct {
	m            sync.Mutex
	sg           []*ec2types.SecurityGroup
	fs           []*efstypes.FileSystemDescription
	mountTargets map[string][]mountTargetItem
}

func filterMatchesTags(tags []ec2types.Tag, filter ec2types.Filter) bool {
	for _, t := range tags {
		tagKey := ptr.Deref(t.Key, "")
		filterName := ptr.Deref(filter.Name, "")
		if tagKey != filterName {
			continue
		}
		tagValue := ptr.Deref(t.Value, "")
		for _, fv := range filter.Values {
			if tagValue == fv {
				return true
			}
		}
	}
	return false
}

func anyFilterMatchTags(tags []ec2types.Tag, filters []ec2types.Filter) bool {
	for _, f := range filters {
		if filterMatchesTags(tags, f) {
			return true
		}
	}
	return false
}

// Config =======

func (s *nfsStore) SetFileSystemLifeCycleState(id string, state efstypes.LifeCycleState) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, fs := range s.fs {
		if *fs.FileSystemId == id {
			fs.LifeCycleState = state
			return
		}
	}
}

func (s *nfsStore) SetFileSystemStatusCapacityById(id string, capacity resource.Quantity) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, fs := range s.fs {
		if *fs.FileSystemId == id {
			fs.SizeInBytes = &efstypes.FileSystemSize{
				Value: capacity.Value(),
			}
			return
		}
	}
}

func (s *nfsStore) GetFileSystemById(id string) *efstypes.FileSystemDescription {
	for _, fs := range s.fs {
		if ptr.Deref(fs.FileSystemId, "") == id {
			return fs
		}
	}
	return nil
}

// Client ===============================

func (s *nfsStore) DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	list := append([]*ec2types.SecurityGroup{}, s.sg...)
	if groupIds != nil {
		list = pie.Filter(list, func(sg *ec2types.SecurityGroup) bool {
			return pie.Contains(groupIds, ptr.Deref(sg.GroupId, ""))
		})
	}
	if filters != nil {
		list = pie.Filter(list, func(sg *ec2types.SecurityGroup) bool {
			return anyFilterMatchTags(sg.Tags, filters)
		})
	}
	result := make([]ec2types.SecurityGroup, 0, len(list))
	for _, x := range list {
		result = append(result, *x)
	}
	return result, nil
}

func (s *nfsStore) CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	tags = append(tags, ec2types.Tag{
		Key:   ptr.To("vpc-id"),
		Value: ptr.To(vpcId),
	})
	sg := &ec2types.SecurityGroup{
		Description: ptr.To(name),
		GroupId:     ptr.To(uuid.NewString()),
		GroupName:   ptr.To(name),
		Tags:        tags,
		VpcId:       ptr.To(vpcId),
	}
	s.sg = append(s.sg, sg)
	return ptr.Deref(sg.GroupId, ""), nil
}

func (s *nfsStore) DeleteSecurityGroup(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	s.sg = pie.Filter(s.sg, func(sg *ec2types.SecurityGroup) bool {
		return ptr.Deref(sg.GroupId, "") != id
	})
	return nil
}

func (s *nfsStore) AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	var securityGroup *ec2types.SecurityGroup
	for _, sg := range s.sg {
		if ptr.Deref(sg.GroupId, "") == groupId {
			securityGroup = sg
			break
		}
	}
	if securityGroup == nil {
		return fmt.Errorf("security group with id %s does not exist", groupId)
	}
	securityGroup.IpPermissions = ipPermissions
	return nil
}

func (s *nfsStore) DescribeFileSystems(ctx context.Context) ([]efstypes.FileSystemDescription, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	result := make([]efstypes.FileSystemDescription, 0, len(s.fs))
	for _, x := range s.fs {
		result = append(result, *x)
	}
	return result, nil
}

func (s *nfsStore) CreateFileSystem(ctx context.Context, performanceMode efstypes.PerformanceMode, throughputMode efstypes.ThroughputMode, tags []efstypes.Tag) (*efs.CreateFileSystemOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	id := uuid.NewString()
	name := awsutil.GetEfsTagValue(tags, "Name")
	if name == "" {
		name = id
	}
	fs := &efstypes.FileSystemDescription{
		FileSystemId:         ptr.To(id),
		LifeCycleState:       efstypes.LifeCycleStateAvailable,
		NumberOfMountTargets: 0,
		PerformanceMode:      performanceMode,
		Tags:                 tags,
		Name:                 ptr.To(name),
		ThroughputMode:       throughputMode,
	}
	s.fs = append(s.fs, fs)

	return &efs.CreateFileSystemOutput{
		CreationTime:                 fs.CreationTime,
		CreationToken:                fs.CreationToken,
		FileSystemId:                 fs.FileSystemId,
		LifeCycleState:               fs.LifeCycleState,
		NumberOfMountTargets:         fs.NumberOfMountTargets,
		OwnerId:                      fs.OwnerId,
		PerformanceMode:              fs.PerformanceMode,
		SizeInBytes:                  fs.SizeInBytes,
		Tags:                         fs.Tags,
		AvailabilityZoneId:           fs.AvailabilityZoneId,
		AvailabilityZoneName:         fs.AvailabilityZoneName,
		Encrypted:                    fs.Encrypted,
		FileSystemArn:                fs.FileSystemArn,
		FileSystemProtection:         fs.FileSystemProtection,
		KmsKeyId:                     fs.KmsKeyId,
		Name:                         fs.Name,
		ProvisionedThroughputInMibps: fs.ProvisionedThroughputInMibps,
		ThroughputMode:               fs.ThroughputMode,
	}, nil
}

func (s *nfsStore) DeleteFileSystem(ctx context.Context, fsId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	s.fs = pie.Filter(s.fs, func(fs *efstypes.FileSystemDescription) bool {
		return ptr.Deref(fs.FileSystemId, "") != fsId
	})
	return nil
}

func (s *nfsStore) DescribeMountTargets(ctx context.Context, fsId string) ([]efstypes.MountTargetDescription, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if s.mountTargets == nil {
		s.mountTargets = map[string][]mountTargetItem{}
	}
	res, ok := s.mountTargets[fsId]
	if !ok {
		return nil, nil
	}
	return pie.Map(res, func(i mountTargetItem) efstypes.MountTargetDescription {
		return i.desc
	}), nil
}

func (s *nfsStore) CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if s.mountTargets == nil {
		s.mountTargets = map[string][]mountTargetItem{}
	}
	list := s.mountTargets[fsId]
	id := uuid.NewString()
	item := mountTargetItem{
		desc: efstypes.MountTargetDescription{
			FileSystemId:   ptr.To(fsId),
			LifeCycleState: efstypes.LifeCycleStateAvailable,
			MountTargetId:  ptr.To(id),
			SubnetId:       ptr.To(subnetId),
			IpAddress:      ptr.To("1.2.3.4"),
		},
		sg: securityGroups,
	}
	list = append(list, item)
	s.mountTargets[fsId] = list
	return id, nil
}

func (s *nfsStore) DeleteMountTarget(ctx context.Context, mountTargetId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if s.mountTargets == nil {
		s.mountTargets = map[string][]mountTargetItem{}
	}
	for fsId, list := range s.mountTargets {
		for _, mt := range list {
			if ptr.Deref(mt.desc.MountTargetId, "") == mountTargetId {
				s.mountTargets[fsId] = pie.Filter(list, func(item mountTargetItem) bool {
					return ptr.Deref(item.desc.MountTargetId, "") != mountTargetId
				})
				return nil
			}
		}
	}
	return nil
}

func (s *nfsStore) DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if s.mountTargets == nil {
		s.mountTargets = map[string][]mountTargetItem{}
	}
	for _, list := range s.mountTargets {
		for _, item := range list {
			if ptr.Deref(item.desc.MountTargetId, "") == mountTargetId {
				return item.sg, nil
			}
		}
	}
	return nil, fmt.Errorf("mount target with id %s does not exist", mountTargetId)
}
