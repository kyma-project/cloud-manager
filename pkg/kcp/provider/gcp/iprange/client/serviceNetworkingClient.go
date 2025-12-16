package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"google.golang.org/api/cloudresourcemanager/v1"

	"github.com/kyma-project/cloud-manager/pkg/composed"

	"google.golang.org/api/option"
	"google.golang.org/api/servicenetworking/v1"
)

// Package client provides GCP API clients for IpRange operations.
//
// HYBRID APPROACH NOTE:
// - ComputeClient: Uses NEW pattern (cloud.google.com/go/compute/apiv1)
// - ServiceNetworkingClient: Uses OLD pattern (google.golang.org/api/servicenetworking/v1)
//
// ServiceNetworkingClient uses the OLD pattern because Google does not provide
// a modern Cloud Client Library for Service Networking API as of December 2024.
// The interface remains clean and testable regardless of underlying implementation.
//
// If cloud.google.com/go/servicenetworking becomes available, migrate to NEW pattern.

type ServiceNetworkingClient interface {
	ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error)
	CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
	// DeleteServiceConnection: Deletes a private service access connection.
	// projectNumber: Project number which is different from project id. Get it by calling client.GetProjectNumber(ctx, projectId)
	DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error)
	PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
	GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error)
}

// NewServiceNetworkingClientProvider creates a GcpClientProvider for ServiceNetworkingClient.
// ServiceNetworking uses OLD pattern (ClientProvider) because Google does not provide
// a modern Cloud Client Library for Service Networking API.
// This wrapper converts OLD pattern to GcpClientProvider interface for consistency.
func NewServiceNetworkingClientProvider() client.GcpClientProvider[ServiceNetworkingClient] {
	// Create the OLD pattern provider
	oldProvider := client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (ServiceNetworkingClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, err
			}
			svcNetClient, err := servicenetworking.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP ServiceNetworking Client: [%w]", err)
			}
			crmService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP CloudResourceManager Service: [%w]", err)
			}
			return NewServiceNetworkingClientForService(svcNetClient, crmService), nil
		},
	)

	// Wrap as GcpClientProvider (just calls oldProvider with hardcoded credentials path)
	return func() ServiceNetworkingClient {
		// Get credentials from GCP config
		credentialsFile := config.GcpConfig.CredentialsFile
		client, err := oldProvider(context.Background(), credentialsFile)
		if err != nil {
			// This should rarely happen since we cache the client.
			// The panic is acceptable here because:
			// 1. Client is cached after first successful creation
			// 2. If credentials are invalid, we want to fail fast at startup
			// 3. This prevents silent failures that would be harder to debug
			panic(fmt.Sprintf("failed to create ServiceNetworking client: %v", err))
		}
		return client
	}
}

// Deprecated: Use NewServiceNetworkingClientProvider instead.
func NewServiceNetworkingClient() client.ClientProvider[ServiceNetworkingClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (ServiceNetworkingClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, err
			}
			svcNetClient, err := servicenetworking.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP ServiceNetworking Client: [%w]", err)
			}
			crmService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP CRM Client: [%w]", err)
			}
			return NewServiceNetworkingClientForService(svcNetClient, crmService), nil
		},
	)
}

func NewServiceNetworkingClientForService(svcNet *servicenetworking.APIService, crmService *cloudresourcemanager.Service) ServiceNetworkingClient {
	return &serviceNetworkingClient{svcNet: svcNet, crmService: crmService}
}

type serviceNetworkingClient struct {
	svcNet     *servicenetworking.APIService
	crmService *cloudresourcemanager.Service
}

func (c *serviceNetworkingClient) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Patch(client.ServiceNetworkingServiceConnectionName, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Force(true).Do()
	client.IncrementCallCounter("ServiceNetworking", "Services.Connections.Patch", "", err)
	logger.Info("PatchServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	ProjectNumber, err := client.GetCachedProjectNumber(projectId, c.crmService)
	if err != nil {
		return nil, err
	}
	network := client.GetVPCPath(strconv.FormatInt(ProjectNumber, 10), vpcId)
	operation, err := c.svcNet.Services.Connections.DeleteConnection(client.ServiceNetworkingServiceConnectionName, &servicenetworking.DeleteConnectionRequest{
		ConsumerNetwork: network,
	}).Do()
	client.IncrementCallCounter("ServiceNetworking", "Services.Connections.DeleteConnection", "", err)
	logger.Info("DeleteServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	out, err := c.svcNet.Services.Connections.List(client.ServiceNetworkingServicePath).Network(network).Do()
	client.IncrementCallCounter("ServiceNetworking", "Services.Connections.List", "", err)
	if err != nil {
		logger.Error(err, "ListServiceConnections", "projectId", projectId, "vpcId", vpcId)
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClient) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Create(client.ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Do()
	client.IncrementCallCounter("ServiceNetworking", "Services.Connections.Create", "", err)
	logger.Info("CreateServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcNet.Operations.Get(operationName).Do()
	client.IncrementCallCounter("ServiceNetworking", "Operations.Get", "", err)
	if err != nil {
		logger.Error(err, "GetServiceNetworkingOperation", "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
