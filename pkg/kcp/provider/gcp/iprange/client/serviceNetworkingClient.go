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

// ServiceNetworkingClient embeds the wrapped client.ServiceNetworkingClient interface.
// The feature-local methods have identical signatures to the wrapped interface,
// so no additional methods are needed.
type ServiceNetworkingClient interface {
	client.ServiceNetworkingClient
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
	return &serviceNetworkingClientAdapter{ServiceNetworkingClient: wrapped}
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

// serviceNetworkingClientAdapter embeds the central client.ServiceNetworkingClient interface.
// Since all methods have identical signatures, the embedded interface provides them directly.
type serviceNetworkingClientAdapter struct {
	client.ServiceNetworkingClient
}

var _ ServiceNetworkingClient = &serviceNetworkingClientAdapter{}

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
