package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	rediscluster "cloud.google.com/go/redis/cluster/apiv1"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

type GcpClients struct {
	ComputeNetworking *compute.NetworksClient
	ComputeRouters    *compute.RoutersClient
	RedisCluster      *rediscluster.CloudRedisClusterClient
}

func NewGcpClients(ctx context.Context, saJsonKeyPath string) (*GcpClients, error) {
	computeNetworking, err := compute.NewNetworksRESTClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
	if err != nil {
		return nil, err
	}

	computeRouters, err := compute.NewRoutersRESTClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
	if err != nil {
		return nil, err
	}

	redisCluster, err := rediscluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
	if err != nil {
		return nil, err
	}

	return &GcpClients{
		ComputeNetworking: computeNetworking,
		ComputeRouters:    computeRouters,
		RedisCluster:      redisCluster,
	}, nil
}

func (c *GcpClients) Close() error {
	var result error
	if c.ComputeNetworking != nil {
		if err := c.ComputeNetworking.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if c.ComputeRouters != nil {
		if err := c.ComputeRouters.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if c.RedisCluster != nil {
		if err := c.RedisCluster.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

type GcpClientProvider[T any] func() T
