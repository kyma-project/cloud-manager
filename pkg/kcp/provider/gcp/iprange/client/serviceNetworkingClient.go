package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/metrics"
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
	return func(_ string) ServiceNetworkingClient {
		return NewServiceNetworkingClientFromWrapped(gcpClients.ServiceNetworkingWrapped())
	}
}

// NewServiceNetworkingClientFromWrapped creates a ServiceNetworkingClient from wrapped interface.
// Used by mock2 for test wiring.
func NewServiceNetworkingClientFromWrapped(wrapped client.ServiceNetworkingClient) ServiceNetworkingClient {
	return &serviceNetworkingClientAdapter{wrapped: wrapped}
}

// NewServiceNetworkingClientProviderV2 creates a ClientProvider (OLD pattern) for v2 legacy code.
// This wraps the clients from GcpClients to avoid duplicate client creation.
func NewServiceNetworkingClientProviderV2(gcpClients *client.GcpClients) client.ClientProvider[ServiceNetworkingClient] {
	return func(ctx context.Context, credentialsFile string) (ServiceNetworkingClient, error) {
		return NewServiceNetworkingClientFromWrapped(gcpClients.ServiceNetworkingWrapped()), nil
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

			httpClient := metrics.NewMetricsHTTPClient(baseClient.Transport)

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
	return &serviceNetworkingClientDirect{svcNet: svcNet, crmService: crmService}
}

// serviceNetworkingClientAdapter wraps the central gcpclient.ServiceNetworkingClient interface
// and delegates to it. Used in the NEW pattern.
type serviceNetworkingClientAdapter struct {
	wrapped client.ServiceNetworkingClient
}

func (c *serviceNetworkingClientAdapter) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return c.wrapped.PatchServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (c *serviceNetworkingClientAdapter) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	return c.wrapped.DeleteServiceConnection(ctx, projectId, vpcId)
}

func (c *serviceNetworkingClientAdapter) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	return c.wrapped.ListServiceConnections(ctx, projectId, vpcId)
}

func (c *serviceNetworkingClientAdapter) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return c.wrapped.CreateServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (c *serviceNetworkingClientAdapter) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	return c.wrapped.GetServiceNetworkingOperation(ctx, operationName)
}

// serviceNetworkingClientDirect is the direct implementation using the legacy API.
// Used by the deprecated NewServiceNetworkingClient and NewServiceNetworkingClientForService.
type serviceNetworkingClientDirect struct {
	svcNet     *servicenetworking.APIService
	crmService *cloudresourcemanager.Service
}

func (c *serviceNetworkingClientDirect) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Patch(client.ServiceNetworkingServiceConnectionName, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Force(true).Do()
	logger.Info("PatchServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClientDirect) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
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

func (c *serviceNetworkingClientDirect) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	out, err := c.svcNet.Services.Connections.List(client.ServiceNetworkingServicePath).Network(network).Do()
	if err != nil {
		logger.Error(err, "ListServiceConnections", "projectId", projectId, "vpcId", vpcId)
		return nil, err
	}
	return out.Connections, nil
}

func (c *serviceNetworkingClientDirect) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	network := client.GetVPCPath(projectId, vpcId)
	operation, err := c.svcNet.Services.Connections.Create(client.ServiceNetworkingServicePath, &servicenetworking.Connection{
		Network:               network,
		ReservedPeeringRanges: reservedIpRanges,
	}).Do()
	logger.Info("CreateServiceConnection", "operation", operation, "err", err)
	return operation, err
}

func (c *serviceNetworkingClientDirect) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcNet.Operations.Get(operationName).Do()
	if err != nil {
		logger.Error(err, "GetServiceNetworkingOperation", "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
