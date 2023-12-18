package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type NetworkClient interface {
	DescribeVpcs(ctx context.Context) ([]types.Vpc, error)
	DescribeSubnets(ctx context.Context, vpcId string) ([]types.Subnet, error)
	CreateSubnet(ctx context.Context, vpcId, az, cidr string) (*types.Subnet, error)
	DeleteSubnet(ctx context.Context, subnetId string) error
}

type EfsClient interface {
}
