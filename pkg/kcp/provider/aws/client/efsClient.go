package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"k8s.io/utils/ptr"
)

type EfsClient interface {
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

func NewEfsClient(svc *efs.Client) EfsClient {
	return &efsClient{svc: svc}
}

var _ EfsClient = (*efsClient)(nil)

type efsClient struct {
	svc *efs.Client
}


func (c *efsClient) DescribeFileSystems(ctx context.Context) ([]efstypes.FileSystemDescription, error) {
	in := &efs.DescribeFileSystemsInput{}
	out, err := c.svc.DescribeFileSystems(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.FileSystems, nil
}

func (c *efsClient) CreateFileSystem(ctx context.Context, performanceMode efstypes.PerformanceMode, throughputMode efstypes.ThroughputMode, tags []efstypes.Tag) (*efs.CreateFileSystemOutput, error) {
	in := &efs.CreateFileSystemInput{
		Encrypted:       aws.Bool(true),
		PerformanceMode: performanceMode,
		Tags:            tags,
		ThroughputMode:  throughputMode,
	}
	out, err := c.svc.CreateFileSystem(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *efsClient) DeleteFileSystem(ctx context.Context, fsId string) error {
	in := &efs.DeleteFileSystemInput{FileSystemId: ptr.To(fsId)}
	_, err := c.svc.DeleteFileSystem(ctx, in)
	return err
}

func (c *efsClient) DescribeMountTargets(ctx context.Context, fsId string) ([]efstypes.MountTargetDescription, error) {
	out, err := c.svc.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: ptr.To(fsId),
	})
	if err != nil {
		return nil, err
	}
	return out.MountTargets, nil
}

func (c *efsClient) CreateMountTarget(ctx context.Context, fsId, subnetId string, securityGroups []string) (string, error) {
	out, err := c.svc.CreateMountTarget(ctx, &efs.CreateMountTargetInput{
		FileSystemId:   ptr.To(fsId),
		SubnetId:       ptr.To(subnetId),
		SecurityGroups: securityGroups,
	})
	if err != nil {
		return "", err
	}
	return ptr.Deref(out.MountTargetId, ""), nil
}

func (c *efsClient) DeleteMountTarget(ctx context.Context, mountTargetId string) error {
	in := &efs.DeleteMountTargetInput{
		MountTargetId: ptr.To(mountTargetId),
	}
	_, err := c.svc.DeleteMountTarget(ctx, in)
	return err
}

func (c *efsClient) DescribeMountTargetSecurityGroups(ctx context.Context, mountTargetId string) ([]string, error) {
	out, err := c.svc.DescribeMountTargetSecurityGroups(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: ptr.To(mountTargetId),
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}
