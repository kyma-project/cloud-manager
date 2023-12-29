package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	awsclient "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/client"
	"k8s.io/utils/pointer"
)

const (
	TcpProtocol = "tcp"
	NfsPort     = int32(2049)
)

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return awsclient.NewCachedSkrClientProvider(
		func(ctx context.Context, region, key, secret, role string) (Client, error) {
			cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
			if err != nil {
				return nil, err
			}
			return newClient(
				ec2.NewFromConfig(cfg),
				efs.NewFromConfig(cfg),
			), nil
		},
	)
}

type Client interface {
	DescribeSecurityGroups(ctx context.Context, filters []ec2Types.Filter, groupIds []string) ([]ec2Types.SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2Types.Tag) (string, error)
	AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2Types.IpPermission) error

	DescribeFileSystems(ctx context.Context) ([]efsTypes.FileSystemDescription, error)
	CreateFileSystem(
		ctx context.Context,
		performanceMode efsTypes.PerformanceMode,
		throughputMode efsTypes.ThroughputMode,
		tags []efsTypes.Tag,
	) (*efs.CreateFileSystemOutput, error)
	DescribeMountTargets(ctx context.Context, fsId string) ([]efsTypes.MountTargetDescription, error)
	CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error)

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

func (c *client) DescribeSecurityGroups(ctx context.Context, filters []ec2Types.Filter, groupIds []string) ([]ec2Types.SecurityGroup, error) {
	out, err := c.ec2Svc.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters:  filters,
		GroupIds: groupIds,
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}

func (c *client) CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2Types.Tag) (string, error) {
	out, err := c.ec2Svc.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: pointer.String(name),
		GroupName:   pointer.String(name),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeSecurityGroup,
				Tags:         tags,
			},
		},
		VpcId: pointer.String(vpcId),
	})
	if err != nil {
		return "", err
	}
	return pointer.StringDeref(out.GroupId, ""), nil
}

func (c *client) AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2Types.IpPermission) error {
	_, err := c.ec2Svc.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       pointer.String(groupId),
		IpPermissions: ipPermissions,
	})
	if err != nil {
		return nil
	}
	return nil
}

func (c *client) DescribeFileSystems(ctx context.Context) ([]efsTypes.FileSystemDescription, error) {
	in := &efs.DescribeFileSystemsInput{}
	out, err := c.efsSvc.DescribeFileSystems(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.FileSystems, nil
}

func (c *client) CreateFileSystem(ctx context.Context, performanceMode efsTypes.PerformanceMode, throughputMode efsTypes.ThroughputMode, tags []efsTypes.Tag) (*efs.CreateFileSystemOutput, error) {
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

func (c *client) DescribeMountTargets(ctx context.Context, fsId string) ([]efsTypes.MountTargetDescription, error) {
	out, err := c.efsSvc.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: pointer.String(fsId),
	})
	if err != nil {
		return nil, err
	}
	return out.MountTargets, nil
}

func (c *client) CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error) {
	out, err := c.efsSvc.CreateMountTarget(ctx, &efs.CreateMountTargetInput{
		FileSystemId:   pointer.String(fsId),
		SubnetId:       pointer.String(subnetId),
		SecurityGroups: securityGroups,
	})
	if err != nil {
		return "", err
	}
	return pointer.StringDeref(out.MountTargetId, ""), nil
}

func (c *client) DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error) {
	out, err := c.efsSvc.DescribeMountTargetSecurityGroups(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: pointer.String(mountTargetId),
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}
