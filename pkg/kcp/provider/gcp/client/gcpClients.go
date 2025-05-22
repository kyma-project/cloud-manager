package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/auth/oauth2adapt"
	compute "cloud.google.com/go/compute/apiv1"
	networkconnectivity "cloud.google.com/go/networkconnectivity/apiv1"
	rediscluster "cloud.google.com/go/redis/cluster/apiv1"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

type GcpClients struct {
	ComputeNetworks                           *compute.NetworksClient
	ComputeAddresses                          *compute.AddressesClient
	ComputeRouters                            *compute.RoutersClient
	ComputeSubnetworks                        *compute.SubnetworksClient
	NetworkConnectivityCrossNetworkAutomation *networkconnectivity.CrossNetworkAutomationClient
	RedisCluster                              *rediscluster.CloudRedisClusterClient
}

func NewGcpClients(ctx context.Context, saJsonKeyPath string, logger logr.Logger) (*GcpClients, error) {
	logger.
		WithValues("saJsonKeyPath", saJsonKeyPath).
		Info("Creating GCP clients")

	b := NewReloadingSaKeyTokenProviderOptionsBuilder(saJsonKeyPath, logger)

	// compute --------------

	computeTokenProvider, err := b.WithScopes(compute.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build compute token provider: %w", err)
	}
	computeTokenSource := oauth2adapt.TokenSourceFromTokenProvider(computeTokenProvider)

	computeNetworks, err := compute.NewNetworksRESTClient(ctx, option.WithTokenSource(computeTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create compute networs client: %w", err)
	}
	computeAddress, err := compute.NewAddressesRESTClient(ctx, option.WithTokenSource(computeTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create compute addresses client: %w", err)
	}
	computeRouters, err := compute.NewRoutersRESTClient(ctx, option.WithTokenSource(computeTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create compute routers client: %w", err)
	}
	computeSubnetworks, err := compute.NewSubnetworksRESTClient(ctx, option.WithTokenSource(computeTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create compute subnetworks client: %w", err)
	}

	// network connectivity ----------------

	networkConnectivityTokenProvider, err := b.WithScopes(networkconnectivity.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build network connectivity token provider: %w", err)
	}
	networkConnectivityTokenSource := oauth2adapt.TokenSourceFromTokenProvider(networkConnectivityTokenProvider)

	ncCrossNetworkAutomation, err := networkconnectivity.NewCrossNetworkAutomationClient(ctx, option.WithTokenSource(networkConnectivityTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create network connectivity cross network automation client: %w", err)
	}

	// redis cluster ----------------

	redisClusterTokenProvider, err := b.WithScopes(rediscluster.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis cluster token provider: %w", err)
	}
	redisClusterTokenSource := oauth2adapt.TokenSourceFromTokenProvider(redisClusterTokenProvider)
	redisCluster, err := rediscluster.NewCloudRedisClusterClient(ctx, option.WithTokenSource(redisClusterTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create redis cluster client: %w", err)
	}

	return &GcpClients{
		ComputeNetworks:    computeNetworks,
		ComputeAddresses:   computeAddress,
		ComputeRouters:     computeRouters,
		ComputeSubnetworks: computeSubnetworks,
		NetworkConnectivityCrossNetworkAutomation: ncCrossNetworkAutomation,
		RedisCluster: redisCluster,
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
