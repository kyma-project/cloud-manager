package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
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
			awsclient.NewEc2Client(ec2.NewFromConfig(cfg)),
			awsclient.NewEfsClient(efs.NewFromConfig(cfg)),
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

func newClient(ec2Client awsclient.Ec2Client, efsClient awsclient.EfsClient) Client {
	return &client{
		Ec2Client: ec2Client,
		EfsClient: efsClient,
	}
}

var _ Client = (*client)(nil)

type client struct {
	awsclient.Ec2Client
	awsclient.EfsClient
}
