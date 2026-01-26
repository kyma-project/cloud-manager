package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/cloudresourcemanager/v1"

	"github.com/kyma-project/cloud-manager/pkg/composed"

	"google.golang.org/api/option"
	"google.golang.org/api/servicenetworking/v1"
)

// Package client provides GCP API clients for IpRange operations.
//
// HYBRID APPROACH NOTE:
// - ComputeClient: Uses NEW pattern (cloud.google.com/go/compute/apiv1)
// - ServiceNetworkingClient: Uses OLD pattern API (google.golang.org/api/servicenetworking/v1)
//
// ServiceNetworkingClient uses the OLD pattern API because Google does not provide
// a modern Cloud Client Library for Service Networking API as of December 2024.
// However, it follows the NEW pattern for dependency injection - clients are initialized
// in GcpClients and injected via main.go, consistent with other GCP controllers.
//
// The interface remains clean and testable regardless of underlying implementation.
// If cloud.google.com/go/servicenetworking becomes available, only the initialization
// in gcpClients.go needs to change; the provider pattern remains the same.

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
// ServiceNetworking uses OLD pattern API (google.golang.org/api/servicenetworking/v1)
// because Google does not provide a modern Cloud Client Library for Service Networking API.
// The clients are initialized in GcpClients and injected here for consistency with other providers.
func NewServiceNetworkingClientProvider(gcpClients *client.GcpClients) client.GcpClientProvider[ServiceNetworkingClient] {
	return func() ServiceNetworkingClient {
		return NewServiceNetworkingClientForService(
			gcpClients.ServiceNetworking,
			gcpClients.CloudResourceManager,
		)
	}
}

// NewServiceNetworkingClientProviderV2 creates a ClientProvider (OLD pattern) for v2 legacy code.
// This wraps the clients from GcpClients to avoid duplicate client creation.
func NewServiceNetworkingClientProviderV2(gcpClients *client.GcpClients) client.ClientProvider[ServiceNetworkingClient] {
	return func(ctx context.Context, credentialsFile string) (ServiceNetworkingClient, error) {
		return NewServiceNetworkingClientForService(
			gcpClients.ServiceNetworking,
			gcpClients.CloudResourceManager,
		), nil
	}
}

// Deprecated: Use NewServiceNetworkingClientProviderV2 instead.
func NewServiceNetworkingClient() client.ClientProvider[ServiceNetworkingClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (ServiceNetworkingClient, error) {
			baseClient, err := client.GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, err
			}

			httpClient := client.NewMetricsHTTPClient(baseClient.Transport)

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
	logger.Info("DeleteServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	out, err := c.svcNet.Services.Connections.List(client.ServiceNetworkingServicePath).Network(network).Do()
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
	logger.Info("CreateServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClient) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcNet.Operations.Get(operationName).Do()
	if err != nil {
		logger.Error(err, "GetServiceNetworkingOperation", "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
