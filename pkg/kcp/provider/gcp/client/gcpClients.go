package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	compute "cloud.google.com/go/compute/apiv1"
	rediscluster "cloud.google.com/go/redis/cluster/apiv1"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

type GcpClients struct {
	ComputeNetworks  *compute.NetworksClient
	ComputeAddresses *compute.AddressesClient
	ComputeRouters   *compute.RoutersClient
	RedisCluster      *rediscluster.CloudRedisClusterClient
}

func NewGcpClients(ctx context.Context, saJsonKeyPath string) (*GcpClients, error) {
	b := NewReloadingSaKeyTokenProviderOptionsBuilder(saJsonKeyPath)

	// compute --------------

	computeHttpClient, err := b.WithScopes(compute.DefaultAuthScopes()).BuildHttpClient()
	if err != nil {
		return nil, fmt.Errorf("create compute http client: %w", err)
	}
	computeNetworks, err := compute.NewNetworksRESTClient(ctx, option.WithHTTPClient(computeHttpClient))
	if err != nil {
		return nil, fmt.Errorf("create compute networs client: %w", err)
	}
	computeAddress, err := compute.NewAddressesRESTClient(ctx, option.WithHTTPClient(computeHttpClient))
	if err != nil {
		return nil, fmt.Errorf("create compute addresses client: %w", err)
	}
	computeRouters, err := compute.NewRoutersRESTClient(ctx, option.WithHTTPClient(computeHttpClient))
	if err != nil {
		return nil, fmt.Errorf("create compute routers client: %w", err)
	}

	//// redis cluster ----------------
	//
	//redisHttpClient, err := b.WithScopes(rediscluster.DefaultAuthScopes()).BuildHttpClient()
	//if err != nil {
	//	return nil, fmt.Errorf("create redis http client: %w", err)
	//}
	//option.WithGRPCDialOption()
	//redisCluster, err := rediscluster.NewCloudRedisClusterClient(ctx, option.WithHTTPClient(redisHttpClient))
	//if err != nil {
	//	return nil, fmt.Errorf("create redis cluster client: %w", err)
	//}

	return &GcpClients{
		ComputeNetworks:  computeNetworks,
		ComputeAddresses: computeAddress,
		ComputeRouters:   computeRouters,
		//RedisCluster:     redisCluster,
	}, nil
}

func (c *GcpClients) Close() error {
	return reflectingClose(c)
}

type GcpClientProvider[T any] func() T

// ReflectingClose ===============================

type closeable interface {
	Close() error
}

func reflectingClose(obj any) error {
	var result error
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("expected a struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.IsZero() {
			continue
		}
		c, ok := field.Interface().(closeable)
		if !ok {
			continue
		}
		if err := c.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
