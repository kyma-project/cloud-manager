package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/auth/oauth2adapt"
	compute "cloud.google.com/go/compute/apiv1"
	filestore "cloud.google.com/go/filestore/apiv1"
	networkconnectivity "cloud.google.com/go/networkconnectivity/apiv1"
	redisinstance "cloud.google.com/go/redis/apiv1"
	rediscluster "cloud.google.com/go/redis/cluster/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	"google.golang.org/grpc"
)

type GcpClients struct {
	ComputeNetworks                           *compute.NetworksClient
	ComputeAddresses                          *compute.AddressesClient
	ComputeGlobalAddresses                    *compute.GlobalAddressesClient // For IpRange global address operations
	ComputeRouters                            *compute.RoutersClient
	ComputeSubnetworks                        *compute.SubnetworksClient
	RegionOperations                          *compute.RegionOperationsClient
	ComputeGlobalOperations                   *compute.GlobalOperationsClient // For IpRange global operation tracking
	NetworkConnectivityCrossNetworkAutomation *networkconnectivity.CrossNetworkAutomationClient
	RedisCluster                              *rediscluster.CloudRedisClusterClient
	RedisInstance                             *redisinstance.CloudRedisClient
	Filestore                                 *filestore.CloudFilestoreManagerClient // For NfsInstance v2
	ServiceNetworking                         *servicenetworking.APIService          // For IpRange PSA connections (OLD pattern API)
	CloudResourceManager                      *cloudresourcemanager.Service          // For IpRange project number lookup (OLD pattern API)
	VpcPeeringClients                         *VpcPeeringClients
}

// The VpcPeeringClients uses a different service account than the other clients and has different permissions as well.
type VpcPeeringClients struct {
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

	computeDialOpts := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "Networks"))),
	}

	computeNetworks, err := compute.NewNetworksRESTClient(ctx, computeDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("create compute networs client: %w", err)
	}

	computeDialOptsAddresses := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "Addresses"))),
	}
	computeAddress, err := compute.NewAddressesRESTClient(ctx, computeDialOptsAddresses...)
	if err != nil {
		return nil, fmt.Errorf("create compute addresses client: %w", err)
	}

	computeDialOptsRouters := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "Routers"))),
	}
	computeRouters, err := compute.NewRoutersRESTClient(ctx, computeDialOptsRouters...)
	if err != nil {
		return nil, fmt.Errorf("create compute routers client: %w", err)
	}

	computeDialOptsSubnetworks := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "Subnetworks"))),
	}
	computeSubnetworks, err := compute.NewSubnetworksRESTClient(ctx, computeDialOptsSubnetworks...)
	if err != nil {
		return nil, fmt.Errorf("create compute subnetworks client: %w", err)
	}

	computeDialOptsRegionOps := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "RegionOperations"))),
	}
	computeRegionOperations, err := compute.NewRegionOperationsRESTClient(ctx, computeDialOptsRegionOps...)
	if err != nil {
		return nil, fmt.Errorf("create compute region operations client: %w", err)
	}

	computeDialOptsGlobalAddr := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "GlobalAddresses"))),
	}
	computeGlobalAddresses, err := compute.NewGlobalAddressesRESTClient(ctx, computeDialOptsGlobalAddr...)
	if err != nil {
		return nil, fmt.Errorf("create compute global addresses client: %w", err)
	}

	computeDialOptsGlobalOps := []option.ClientOption{
		option.WithTokenSource(computeTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "GlobalOperations"))),
	}
	computeGlobalOperations, err := compute.NewGlobalOperationsRESTClient(ctx, computeDialOptsGlobalOps...)
	if err != nil {
		return nil, fmt.Errorf("create compute global operations client: %w", err)
	}

	// network connectivity ----------------

	networkConnectivityTokenProvider, err := b.WithScopes(networkconnectivity.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build network connectivity token provider: %w", err)
	}
	networkConnectivityTokenSource := oauth2adapt.TokenSourceFromTokenProvider(networkConnectivityTokenProvider)

	ncDialOpts := []option.ClientOption{
		option.WithTokenSource(networkConnectivityTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("NetworkConnectivity", "CrossNetworkAutomation"))),
	}
	ncCrossNetworkAutomation, err := networkconnectivity.NewCrossNetworkAutomationClient(ctx, ncDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("create network connectivity cross network automation client: %w", err)
	}

	// redis cluster ----------------

	redisClusterTokenProvider, err := b.WithScopes(rediscluster.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis cluster token provider: %w", err)
	}
	redisClusterTokenSource := oauth2adapt.TokenSourceFromTokenProvider(redisClusterTokenProvider)
	redisClusterDialOpts := []option.ClientOption{
		option.WithTokenSource(redisClusterTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("RedisCluster", "CloudRedisCluster"))),
	}
	redisCluster, err := rediscluster.NewCloudRedisClusterClient(ctx, redisClusterDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("create redis cluster client: %w", err)
	}

	// redis instance ----------------
	redisInstanceTokenProvider, err := b.WithScopes(redisinstance.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis instance token provider: %w", err)
	}
	redisInstanceTokenSource := oauth2adapt.TokenSourceFromTokenProvider(redisInstanceTokenProvider)
	redisInstanceDialOpts := []option.ClientOption{
		option.WithTokenSource(redisInstanceTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("RedisInstance", "CloudRedis"))),
	}
	redisInstance, err := redisinstance.NewCloudRedisClient(ctx, redisInstanceDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("create redis instance client: %w", err)
	}

	// filestore ----------------
	filestoreTokenProvider, err := b.WithScopes(filestore.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create filestore token provider: %w", err)
	}
	filestoreTokenSource := oauth2adapt.TokenSourceFromTokenProvider(filestoreTokenProvider)
	filestoreClient, err := filestore.NewCloudFilestoreManagerClient(ctx, option.WithTokenSource(filestoreTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create filestore client: %w", err)
	}

	// service networking and cloud resource manager ----------------
	// ServiceNetworking uses OLD pattern API (google.golang.org/api/servicenetworking/v1)
	// because Google does not provide a modern Cloud Client Library for Service Networking API
	serviceNetworkingTokenProvider, err := b.WithScopes([]string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/service.management",
	}).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build service networking token provider: %w", err)
	}
	serviceNetworkingTokenSource := oauth2adapt.TokenSourceFromTokenProvider(serviceNetworkingTokenProvider)

	serviceNetworking, err := servicenetworking.NewService(ctx, option.WithTokenSource(serviceNetworkingTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create service networking client: %w", err)
	}
	cloudResourceManager, err := cloudresourcemanager.NewService(ctx, option.WithTokenSource(serviceNetworkingTokenSource))
	if err != nil {
		return nil, fmt.Errorf("create cloud resource manager client: %w", err)
	}

	// vpc peering clients ----------------
	// Compute networks client for VPC peering, uses a different service account
	vpcPeeringComputeNetworksTokenProvider, err := vpcPeeringClientBuilder.WithScopes(compute.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build vpc peering compute token provider: %w", err)
	}
	vpcPeeringComputeNetworksTokenSource := oauth2adapt.TokenSourceFromTokenProvider(vpcPeeringComputeNetworksTokenProvider)
	vpcPeeringNetworksDialOpts := []option.ClientOption{
		option.WithTokenSource(vpcPeeringComputeNetworksTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("Compute", "VpcPeeringNetworks"))),
	}
	vpcPeeringComputeNetworks, err := compute.NewNetworksRESTClient(ctx, vpcPeeringNetworksDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating vpc peering compute networks client: %w", err)
	}
	// resource manager client for VPC peering, uses a different service account----------------

	vpcPeeringResourceManagerTokenProvider, err := vpcPeeringClientBuilder.WithScopes(resourcemanager.DefaultAuthScopes()).BuildTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to build vpc peering resource manager token provider: %w", err)
	}
	vpcPeeringresourceManagerTokenSource := oauth2adapt.TokenSourceFromTokenProvider(vpcPeeringResourceManagerTokenProvider)
	vpcPeeringTagBindingsDialOpts := []option.ClientOption{
		option.WithTokenSource(vpcPeeringresourceManagerTokenSource),
		option.WithGRPCDialOption(grpc.WithUnaryInterceptor(UnaryClientInterceptor("ResourceManager", "TagBindings"))),
	}
	vpcPeeringresourceManagerTagBindings, err := resourcemanager.NewTagBindingsRESTClient(ctx, vpcPeeringTagBindingsDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating resource_manager tag bindings client: %w", err)
	}

	return &GcpClients{
		ComputeNetworks:                           computeNetworks,
		ComputeAddresses:                          computeAddress,
		ComputeGlobalAddresses:                    computeGlobalAddresses,
		ComputeRouters:                            computeRouters,
		ComputeSubnetworks:                        computeSubnetworks,
		RegionOperations:                          computeRegionOperations,
		ComputeGlobalOperations:                   computeGlobalOperations,
		NetworkConnectivityCrossNetworkAutomation: ncCrossNetworkAutomation,
		RedisCluster:                              redisCluster,
		RedisInstance:                             redisInstance,
		Filestore:                                 filestoreClient,
		ServiceNetworking:                         serviceNetworking,
		CloudResourceManager:                      cloudResourceManager,
		VpcPeeringClients: &VpcPeeringClients{
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
