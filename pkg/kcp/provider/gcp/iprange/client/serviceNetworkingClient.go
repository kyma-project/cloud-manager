package client

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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

// serviceNetworkingClientAdapter embeds the central client.ServiceNetworkingClient interface.
// Since all methods have identical signatures, the embedded interface provides them directly.
type serviceNetworkingClientAdapter struct {
	client.ServiceNetworkingClient
}

var _ ServiceNetworkingClient = &serviceNetworkingClientAdapter{}
