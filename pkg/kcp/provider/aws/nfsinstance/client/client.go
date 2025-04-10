package client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

const (
	TcpProtocol = "tcp"
	NfsPort     = int32(2049)
)

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(
			ec2.NewFromConfig(cfg),
			efs.NewFromConfig(cfg),
		), nil
	}
}

type Client interface {
	DescribeSubnet(ctx context.Context, subnetId string) (*ec2types.Subnet, error)
	DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error)
	DeleteSecurityGroup(ctx context.Context, id string) error
	AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error

	DescribeFileSystems(ctx context.Context) ([]efstypes.FileSystemDescription, error)
	CreateFileSystem(
		ctx context.Context,
		performanceMode efstypes.PerformanceMode,
		throughputMode efstypes.ThroughputMode,
		tags []efstypes.Tag,
	) (*efs.CreateFileSystemOutput, error)
	DeleteFileSystem(ctx context.Context, fsId string) error
	DescribeMountTargets(ctx context.Context, fsId string) ([]efstypes.MountTargetDescription, error)
	CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error)
	DeleteMountTarget(ctx context.Context, mountTargetId string) error

	DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error)
}

func newClient(ec2Svc *ec2.Client, efsSvc *efs.Client) Client {
	return &client{
		ec2Svc: ec2Svc,
		efsSvc: efsSvc,
	}
}

type client struct {
	ec2Svc *ec2.Client
	efsSvc *efs.Client
}

func (c *client) DescribeSubnet(ctx context.Context, subnetId string) (*ec2types.Subnet, error) {
	out, err := c.ec2Svc.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   ptr.To("subnet-id"),
				Values: []string{subnetId},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Subnets) > 1 {
		return nil, fmt.Errorf("expected at most one subnet by id, but got: %v", pie.Map(out.Subnets, func(s ec2types.Subnet) string {
			return ptr.Deref(s.SubnetId, "")
		}))
	}
	var result *ec2types.Subnet
	if len(out.Subnets) > 0 {
		result = &out.Subnets[0]
	}
	return result, nil
}

func (c *client) DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error) {
	out, err := c.ec2Svc.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters:  filters,
		GroupIds: groupIds,
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}

func (c *client) CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error) {
	out, err := c.ec2Svc.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: ptr.To(name),
		GroupName:   ptr.To(name),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSecurityGroup,
				Tags:         tags,
			},
		},
		VpcId: ptr.To(vpcId),
	})
	if err != nil {
		return "", err
	}
	return ptr.Deref(out.GroupId, ""), nil
}

func (c *client) DeleteSecurityGroup(ctx context.Context, id string) error {
	in := &ec2.DeleteSecurityGroupInput{
		GroupId: ptr.To(id),
	}
	_, err := c.ec2Svc.DeleteSecurityGroup(ctx, in)
	return err
}

func (c *client) AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error {
	_, err := c.ec2Svc.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       ptr.To(groupId),
		IpPermissions: ipPermissions,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *client) DescribeFileSystems(ctx context.Context) ([]efstypes.FileSystemDescription, error) {
	in := &efs.DescribeFileSystemsInput{}
	out, err := c.efsSvc.DescribeFileSystems(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.FileSystems, nil
}

func (c *client) CreateFileSystem(ctx context.Context, performanceMode efstypes.PerformanceMode, throughputMode efstypes.ThroughputMode, tags []efstypes.Tag) (*efs.CreateFileSystemOutput, error) {
	in := &efs.CreateFileSystemInput{
		Encrypted:       aws.Bool(true),
		PerformanceMode: performanceMode,
		Tags:            tags,
		ThroughputMode:  throughputMode,
	}
	out, err := c.efsSvc.CreateFileSystem(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) DeleteFileSystem(ctx context.Context, fsId string) error {
	in := &efs.DeleteFileSystemInput{FileSystemId: ptr.To(fsId)}
	_, err := c.efsSvc.DeleteFileSystem(ctx, in)
	return err
}

func (c *client) DescribeMountTargets(ctx context.Context, fsId string) ([]efstypes.MountTargetDescription, error) {
	out, err := c.efsSvc.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: ptr.To(fsId),
	})
	if err != nil {
		return nil, err
	}
	return out.MountTargets, nil
}

func (c *client) CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error) {
	out, err := c.efsSvc.CreateMountTarget(ctx, &efs.CreateMountTargetInput{
		FileSystemId:   ptr.To(fsId),
		SubnetId:       ptr.To(subnetId),
		SecurityGroups: securityGroups,
	})
	if err != nil {
		return "", err
	}
	return ptr.Deref(out.MountTargetId, ""), nil
}

func (c *client) DeleteMountTarget(ctx context.Context, mountTargetId string) error {
	in := &efs.DeleteMountTargetInput{
		MountTargetId: ptr.To(mountTargetId),
	}
	_, err := c.efsSvc.DeleteMountTarget(ctx, in)
	return err
}

func (c *client) DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error) {
	out, err := c.efsSvc.DescribeMountTargetSecurityGroups(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: ptr.To(mountTargetId),
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}
