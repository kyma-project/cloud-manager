package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type CreateServiceConnectionPolicyRequest struct {
	ProjectId     string
	Region        string
	Name          string
	Network       string
	Subnets       []string
	IdempotenceId string
}

// NetworkConnectivityClient embeds the wrapped gcpclient.NetworkConnectivityClient interface
// and adds the CreateServiceConnectionPolicyForRedis method which contains real business logic
// (VPC path construction, parent/name building, hardcoded serviceClass and description, PSC config).
// Actions call the wrapped methods directly for Get, Update, and Delete operations.
type NetworkConnectivityClient interface {
	gcpclient.NetworkConnectivityClient

	// CreateServiceConnectionPolicyForRedis creates a service connection policy configured for Redis Cluster.
	CreateServiceConnectionPolicyForRedis(ctx context.Context, request CreateServiceConnectionPolicyRequest) error
}

func NewNetworkConnectivityClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[NetworkConnectivityClient] {
	return func(_ string) NetworkConnectivityClient {
		return NewNetworkConnectivityClient(gcpClients)
	}
}

func NewNetworkConnectivityClient(gcpClients *gcpclient.GcpClients) NetworkConnectivityClient {
	return NewNetworkConnectivityClientFromWrapped(gcpClients.NetworkConnectivityWrapped())
}

func NewNetworkConnectivityClientFromWrapped(ncClient gcpclient.NetworkConnectivityClient) NetworkConnectivityClient {
	return &networkConnectivityClient{NetworkConnectivityClient: ncClient}
}

type networkConnectivityClient struct {
	gcpclient.NetworkConnectivityClient
}

var _ NetworkConnectivityClient = &networkConnectivityClient{}

func (ncClient *networkConnectivityClient) CreateServiceConnectionPolicyForRedis(ctx context.Context, request CreateServiceConnectionPolicyRequest) error {
	networkNameFull := fmt.Sprintf("projects/%s/global/networks/%s", request.ProjectId, request.Network)
	parent := fmt.Sprintf("projects/%s/locations/%s", request.ProjectId, request.Region)
	connectionPolicyNameFull := fmt.Sprintf("%s/serviceConnectionPolicies/%s", parent, request.Name)

	_, err := ncClient.NetworkConnectivityClient.CreateServiceConnectionPolicy(ctx, &networkconnectivitypb.CreateServiceConnectionPolicyRequest{
		Parent:                    parent,
		ServiceConnectionPolicyId: request.Name,
		ServiceConnectionPolicy: &networkconnectivitypb.ServiceConnectionPolicy{
			Name:         connectionPolicyNameFull,
			Network:      networkNameFull,
			ServiceClass: "gcp-memorystore-redis",
			Description:  "Managed by cloud-manager. Used for Redis Cluster.",
			PscConfig: &networkconnectivitypb.ServiceConnectionPolicy_PscConfig{
				Subnetworks:              request.Subnets,
				ProducerInstanceLocation: networkconnectivitypb.ServiceConnectionPolicy_PscConfig_PRODUCER_INSTANCE_LOCATION_UNSPECIFIED,
			},
		},
		RequestId: request.IdempotenceId,
	})

	if err != nil {
		return err
	}

	return nil
}
