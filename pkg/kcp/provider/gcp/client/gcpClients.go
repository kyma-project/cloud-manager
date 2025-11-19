package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/auth/oauth2adapt"
	compute "cloud.google.com/go/compute/apiv1"
	networkconnectivity "cloud.google.com/go/networkconnectivity/apiv1"
	redisinstance "cloud.google.com/go/redis/apiv1"
	rediscluster "cloud.google.com/go/redis/cluster/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

type GcpClients struct {
	ComputeNetworks                           *compute.NetworksClient
	ComputeAddresses                          *compute.AddressesClient
	ComputeRouters                            *compute.RoutersClient
	ComputeSubnetworks                        *compute.SubnetworksClient
	RegionOperations                          *compute.RegionOperationsClient
	NetworkConnectivityCrossNetworkAutomation *networkconnectivity.CrossNetworkAutomationClient
	RedisCluster                              *rediscluster.CloudRedisClusterClient
	RedisInstance                             *redisinstance.CloudRedisClient
	VpcPeeringClients                         *VpcPeeringClients
}

// The VpcPeeringClients uses a different service account than the other clients and has different permissions as well.
type VpcPeeringClients struct {
	ComputeGlobalOperations    *compute.GlobalOperationsClient
	ComputeNetworks            *compute.NetworksClient
	ResourceManagerTagBindings *resourcemanager.TagBindingsClient
}

func NewGcpClients(ctx context.Context, credentialsFile string, peeringCredentialsFile string, logger logr.Logger) (*GcpClients, error) {
	if credentialsFile == "" || credentialsFile == "none" || peeringCredentialsFile == "" || peeringCredentialsFile == "none" {
		logger.Info("Creating GCP clients stub since no GCP credentials provided")
		return &GcpClients{}, nil
	}

	logger.
		WithValues("credentialsFile", credentialsFile).
		WithValues("peeringCredentialsFile", peeringCredentialsFile).
		Info("Creating GCP clients")

	b := NewReloadingSaKeyTokenProviderOptionsBuilder(credentialsFile, logger)
	vpcPeeringClientBuilder := NewReloadingSaKeyTokenProviderOptionsBuilder(peeringCredentialsFile, logger)

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
	computeRegionOperations, err := compute.NewRegionOperationsRESTClient(ctx, option.WithTokenSource(computeTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create compute region operations client: %w", err)
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

	// redis instance ----------------
	redisInstanceTokenProvider, err := b.WithScopes(redisinstance.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis instance token provider: %w", err)
	}
	redisInstanceTokenSource := oauth2adapt.TokenSourceFromTokenProvider(redisInstanceTokenProvider)
	redisInstance, err := redisinstance.NewCloudRedisClient(ctx, option.WithTokenSource(redisInstanceTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create redis instance client: %w", err)
	}

	// vpc peering clients ----------------
	// Compute networks client for VPC peering, uses a different service account
	vpcPeeringComputeNetworksTokenProvider, err := vpcPeeringClientBuilder.WithScopes(compute.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build vpc peering compute token provider: %w", err)
	}
	vpcPeeringComputeNetworksTokenSource := oauth2adapt.TokenSourceFromTokenProvider(vpcPeeringComputeNetworksTokenProvider)
	vpcPeeringComputeNetworks, err := compute.NewNetworksRESTClient(ctx, option.WithTokenSource(vpcPeeringComputeNetworksTokenSource))
	if err != nil {
		return nil, fmt.Errorf("error creating vpc peering compute networks client: %w", err)
	}
	vpcPeeringComputeGlobalOperations, err := compute.NewGlobalOperationsRESTClient(ctx, option.WithTokenSource(vpcPeeringComputeNetworksTokenSource))
	if err != nil {
		return nil, fmt.Errorf("error creating vpc peering compute operations client: %w", err)
	}
	// resource manager client for VPC peering, uses a different service account----------------
	vpcPeeringResourceManagerTokenProvider, err := vpcPeeringClientBuilder.WithScopes(resourcemanager.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build vpc peering resource manager token provider: %w", err)
	}
	vpcPeeringresourceManagerTokenSource := oauth2adapt.TokenSourceFromTokenProvider(vpcPeeringResourceManagerTokenProvider)
	vpcPeeringresourceManagerTagBindings, err := resourcemanager.NewTagBindingsRESTClient(ctx, option.WithTokenSource(vpcPeeringresourceManagerTokenSource))
	if err != nil {
		return nil, fmt.Errorf("error creating resource_manager tag bindings client: %w", err)
	}

	return &GcpClients{
		ComputeNetworks:    computeNetworks,
		ComputeAddresses:   computeAddress,
		ComputeRouters:     computeRouters,
		ComputeSubnetworks: computeSubnetworks,
		RegionOperations:   computeRegionOperations,
		NetworkConnectivityCrossNetworkAutomation: ncCrossNetworkAutomation,
		RedisCluster:  redisCluster,
		RedisInstance: redisInstance,
		VpcPeeringClients: &VpcPeeringClients{
			ComputeGlobalOperations:    vpcPeeringComputeGlobalOperations,
			ComputeNetworks:            vpcPeeringComputeNetworks,
			ResourceManagerTagBindings: vpcPeeringresourceManagerTagBindings,
		},
	}, nil
}

func (c *GcpClients) Close() error {
	return reflectingClose(c)
}

func (c *VpcPeeringClients) Close() error {
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
