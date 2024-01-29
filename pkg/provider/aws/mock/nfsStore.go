package mock

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"k8s.io/utils/pointer"
)

type NfsConfig interface {
}

type mountTargetItem struct {
	desc efsTypes.MountTargetDescription
	sg   []string
}

type nfsStore struct {
	sg           []ec2Types.SecurityGroup
	fs           []efsTypes.FileSystemDescription
	mountTargets map[string][]mountTargetItem
}

func filterMatchesTags(tags []ec2Types.Tag, filter ec2Types.Filter) bool {
	for _, t := range tags {
		tagKey := pointer.StringDeref(t.Key, "")
		filterName := pointer.StringDeref(filter.Name, "")
		if tagKey != filterName {
			continue
		}
		tagValue := pointer.StringDeref(t.Value, "")
		for _, fv := range filter.Values {
			if tagValue == fv {
				return true
			}
		}
	}
	return false
}

func anyFilterMatchTags(tags []ec2Types.Tag, filters []ec2Types.Filter) bool {
	for _, f := range filters {
		if filterMatchesTags(tags, f) {
			return true
		}
	}
	return false
}

func (s *nfsStore) DescribeSecurityGroups(ctx context.Context, filters []ec2Types.Filter, groupIds []string) ([]ec2Types.SecurityGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	result := append([]ec2Types.SecurityGroup{}, s.sg...)
	if groupIds != nil {
		result = pie.Filter(result, func(sg ec2Types.SecurityGroup) bool {
			return pie.Contains(groupIds, pointer.StringDeref(sg.GroupId, ""))
		})
	}
	if filters != nil {
		result = pie.Filter(result, func(sg ec2Types.SecurityGroup) bool {
			return anyFilterMatchTags(sg.Tags, filters)
		})
	}
	return result, nil
}

func (s *nfsStore) CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2Types.Tag) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	tags = append(tags, ec2Types.Tag{
		Key:   pointer.String("vpc-id"),
		Value: pointer.String(vpcId),
	})
	sg := ec2Types.SecurityGroup{
		Description: pointer.String(name),
		GroupId:     pointer.String(uuid.NewString()),
		GroupName:   pointer.String(name),
		Tags:        tags,
		VpcId:       pointer.String(vpcId),
	}
	s.sg = append(s.sg, sg)
	return pointer.StringDeref(sg.GroupId, ""), nil
}

func (s *nfsStore) AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2Types.IpPermission) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	var securityGroup *ec2Types.SecurityGroup
	for _, sg := range s.sg {
		if pointer.StringDeref(sg.GroupId, "") == groupId {
			securityGroup = &sg
			break
		}
	}
	if securityGroup == nil {
		return fmt.Errorf("security group with id %s does not exist", groupId)
	}
	securityGroup.IpPermissions = ipPermissions
	return nil
}

func (s *nfsStore) DescribeFileSystems(ctx context.Context) ([]efsTypes.FileSystemDescription, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	return s.fs, nil
}

func (s *nfsStore) CreateFileSystem(ctx context.Context, performanceMode efsTypes.PerformanceMode, throughputMode efsTypes.ThroughputMode, tags []efsTypes.Tag) (*efs.CreateFileSystemOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	id := uuid.NewString()
	fs := efsTypes.FileSystemDescription{
		FileSystemId:         pointer.String(id),
		LifeCycleState:       efsTypes.LifeCycleStateAvailable,
		NumberOfMountTargets: 0,
		PerformanceMode:      performanceMode,
		Tags:                 tags,
		Name:                 pointer.String(id),
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

func (s *nfsStore) DescribeMountTargets(ctx context.Context, fsId string) ([]efsTypes.MountTargetDescription, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	res, ok := s.mountTargets[fsId]
	if !ok {
		return nil, nil
	}
	return pie.Map(res, func(i mountTargetItem) efsTypes.MountTargetDescription {
		return i.desc
	}), nil
}

func (s *nfsStore) CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	list := s.mountTargets[fsId]
	id := uuid.NewString()
	item := mountTargetItem{
		desc: efsTypes.MountTargetDescription{
			FileSystemId:   pointer.String(fsId),
			LifeCycleState: efsTypes.LifeCycleStateAvailable,
			MountTargetId:  pointer.String(id),
			SubnetId:       pointer.String(subnetId),
			IpAddress:      pointer.String("1.2.3.4"),
		},
		sg: securityGroups,
	}
	list = append(list, item)
	s.mountTargets[fsId] = list
	return id, nil
}

func (s *nfsStore) DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	for _, list := range s.mountTargets {
		for _, item := range list {
			if pointer.StringDeref(item.desc.MountTargetId, "") == mountTargetId {
				return item.sg, nil
			}
		}
	}
	return nil, fmt.Errorf("mount target with id %s does not exist", mountTargetId)
}
